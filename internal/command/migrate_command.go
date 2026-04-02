package command

import (
	"context"
	"fmt"
	"sort"
	"time"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	"github.com/spf13/cobra"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Migration represents a single migration step.
type Migration struct {
	Version     string
	Description string
	Up          func(ctx context.Context, db *mongo.Database) error
}

// migrationRecord tracks which migrations have been applied.
type migrationRecord struct {
	Version     string    `bson:"version"`
	Description string    `bson:"description"`
	AppliedAt   time.Time `bson:"appliedAt"`
}

const migrationsCollection = "migrations"

// registry holds all registered migrations in order.
var registry []Migration

// RegisterMigration adds a migration to the registry. Call this from init() in migration files.
func RegisterMigration(m Migration) {
	registry = append(registry, m)
}

func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
	}

	cmd.AddCommand(
		newMigrateUpCommand(),
		newMigrateRunCommand(),
		newMigrateListCommand(),
		newMigrateStatusCommand(),
	)

	return cmd
}

// migrate up - Run all pending migrations
func newMigrateUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			l := logger.FromContext(ctx)

			cmdCtx, err := NewCommandContext()
			if err != nil {
				return fmt.Errorf("failed to create command context: %w", err)
			}
			defer cmdCtx.Close()

			db := cmdCtx.GetMongoDatabase()
			applied, err := getAppliedMigrations(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to get applied migrations: %w", err)
			}

			migrations := getSortedMigrations()
			pending := 0

			for _, m := range migrations {
				if applied[m.Version] {
					continue
				}

				l.Info(fmt.Sprintf("Running migration: %s - %s", m.Version, m.Description),
					logging.String("version", m.Version),
				)

				if err := m.Up(ctx, db); err != nil {
					return fmt.Errorf("migration %s failed: %w", m.Version, err)
				}

				// Record migration
				_, err := db.Collection(migrationsCollection).InsertOne(ctx, migrationRecord{
					Version:     m.Version,
					Description: m.Description,
					AppliedAt:   time.Now(),
				})
				if err != nil {
					return fmt.Errorf("failed to record migration %s: %w", m.Version, err)
				}

				l.Info(fmt.Sprintf("Migration %s completed successfully", m.Version),
					logging.String("version", m.Version),
				)
				pending++
			}

			if pending == 0 {
				l.Info("No pending migrations")
			} else {
				l.Info(fmt.Sprintf("Applied %d migration(s)", pending),
					logging.Int("count", pending),
				)
			}

			return nil
		},
	}
}

// migrate run <version> - Run a specific migration by version
func newMigrateRunCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "run [version]",
		Short: "Run a specific migration by version",
		Long:  "Run a single migration identified by its version string. Use --force to re-run an already applied migration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			ctx := context.Background()
			l := logger.FromContext(ctx)

			// Find migration in registry
			migrations := getSortedMigrations()
			var target *Migration
			for i := range migrations {
				if migrations[i].Version == version {
					target = &migrations[i]
					break
				}
			}
			if target == nil {
				fmt.Println("\nAvailable migrations:")
				for _, m := range migrations {
					fmt.Printf("  %s - %s\n", m.Version, m.Description)
				}
				return fmt.Errorf("migration '%s' not found", version)
			}

			cmdCtx, err := NewCommandContext()
			if err != nil {
				return fmt.Errorf("failed to create command context: %w", err)
			}
			defer cmdCtx.Close()

			db := cmdCtx.GetMongoDatabase()
			applied, err := getAppliedMigrations(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to get applied migrations: %w", err)
			}

			if applied[version] && !force {
				return fmt.Errorf("migration '%s' already applied. Use --force to re-run", version)
			}

			l.Info(fmt.Sprintf("Running migration: %s - %s", target.Version, target.Description),
				logging.String("version", target.Version),
				logging.Bool("force", force),
			)

			if err := target.Up(ctx, db); err != nil {
				return fmt.Errorf("migration %s failed: %w", target.Version, err)
			}

			// Record migration (skip if already recorded)
			if !applied[version] {
				_, err := db.Collection(migrationsCollection).InsertOne(ctx, migrationRecord{
					Version:     target.Version,
					Description: target.Description,
					AppliedAt:   time.Now(),
				})
				if err != nil {
					return fmt.Errorf("failed to record migration %s: %w", target.Version, err)
				}
			}

			l.Info(fmt.Sprintf("Migration %s completed successfully", target.Version),
				logging.String("version", target.Version),
			)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Re-run migration even if already applied")
	return cmd
}

// migrate list - List all available migrations
func newMigrateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all available migration versions",
		Run: func(cmd *cobra.Command, args []string) {
			migrations := getSortedMigrations()
			fmt.Println("\nAvailable migrations:")
			fmt.Println("─────────────────────────────────────────────────")
			for _, m := range migrations {
				fmt.Printf("  %s - %s\n", m.Version, m.Description)
			}
			fmt.Println("─────────────────────────────────────────────────")
			fmt.Printf("\nTotal: %d migration(s)\n", len(migrations))
		},
	}
}

// migrate status - Show migration status
func newMigrateStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status (applied vs pending)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			l := logger.FromContext(ctx)

			cmdCtx, err := NewCommandContext()
			if err != nil {
				return fmt.Errorf("failed to create command context: %w", err)
			}
			defer cmdCtx.Close()

			db := cmdCtx.GetMongoDatabase()
			applied, err := getAppliedMigrations(ctx, db)
			if err != nil {
				return fmt.Errorf("failed to get applied migrations: %w", err)
			}

			migrations := getSortedMigrations()

			fmt.Println("Migration Status:")
			fmt.Println("─────────────────────────────────────────────────")
			for _, m := range migrations {
				status := "PENDING"
				if applied[m.Version] {
					status = "APPLIED"
				}
				fmt.Printf("  [%s] %s - %s\n", status, m.Version, m.Description)
			}
			fmt.Println("─────────────────────────────────────────────────")

			pendingCount := 0
			for _, m := range migrations {
				if !applied[m.Version] {
					pendingCount++
				}
			}

			l.Info("Migration status checked",
				logging.Int("total", len(migrations)),
				logging.Int("applied", len(migrations)-pendingCount),
				logging.Int("pending", pendingCount),
			)

			return nil
		},
	}
}

func getAppliedMigrations(ctx context.Context, db *mongo.Database) (map[string]bool, error) {
	collection := db.Collection(migrationsCollection)

	// Ensure unique index on version
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create migrations index: %w", err)
	}

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer cursor.Close(ctx)

	applied := make(map[string]bool)
	for cursor.Next(ctx) {
		var record migrationRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, fmt.Errorf("failed to decode migration record: %w", err)
		}
		applied[record.Version] = true
	}
	return applied, nil
}

func getSortedMigrations() []Migration {
	sorted := make([]Migration, len(registry))
	copy(sorted, registry)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})
	return sorted
}
