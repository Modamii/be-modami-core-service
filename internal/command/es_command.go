package command

import (
	"context"
	"fmt"

	"be-modami-core-service/internal/adapter/repository"
	"be-modami-core-service/internal/domain"
	es "be-modami-core-service/pkg/elasticsearch"

	logging "gitlab.com/lifegoeson-libs/pkg-logging"
	"gitlab.com/lifegoeson-libs/pkg-logging/logger"

	"github.com/spf13/cobra"
)

func NewESCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "es",
		Short: "Elasticsearch management commands",
	}

	cmd.AddCommand(
		newESInitIndexCommand(),
		newESDeleteIndexCommand(),
		newESReindexCommand(),
		newESHealthCommand(),
	)

	return cmd
}

// es init-index - Create/initialize Elasticsearch products index
func newESInitIndexCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init-index",
		Short: "Initialize Elasticsearch products index with mappings",
		Long:  "Creates the products index with proper mappings and analyzers if it doesn't exist",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			l := logger.FromContext(ctx)

			cmdCtx, err := NewCommandContext()
			if err != nil {
				return fmt.Errorf("failed to create command context: %w", err)
			}
			defer cmdCtx.Close()

			if cmdCtx.ESClient == nil {
				return fmt.Errorf("elasticsearch is not available")
			}

			if err := cmdCtx.ESClient.EnsureProductIndices(ctx); err != nil {
				return fmt.Errorf("failed to initialize products index: %w", err)
			}

			l.Info("Elasticsearch products index initialized successfully")
			return nil
		},
	}
}

// es delete-index - Delete Elasticsearch products index
func newESDeleteIndexCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-index",
		Short: "Delete Elasticsearch products index",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			l := logger.FromContext(ctx)

			cmdCtx, err := NewCommandContext()
			if err != nil {
				return fmt.Errorf("failed to create command context: %w", err)
			}
			defer cmdCtx.Close()

			if cmdCtx.ESClient == nil {
				return fmt.Errorf("elasticsearch is not available")
			}

			if err := cmdCtx.ESClient.DeleteProductIndices(ctx); err != nil {
				return fmt.Errorf("failed to delete products index: %w", err)
			}

			l.Info("Elasticsearch products index deleted successfully")
			return nil
		},
	}
}

// es reindex - Full reindex of all active products from MongoDB to ES
func newESReindexCommand() *cobra.Command {
	var batchSize int
	var clean bool

	cmd := &cobra.Command{
		Use:   "reindex",
		Short: "Reindex all active products from MongoDB to Elasticsearch",
		Long:  "Scans all active products in MongoDB and bulk indexes them into Elasticsearch. Use --clean to delete the index before reindexing.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			l := logger.FromContext(ctx)

			cmdCtx, err := NewCommandContext()
			if err != nil {
				return fmt.Errorf("failed to create command context: %w", err)
			}
			defer cmdCtx.Close()

			if cmdCtx.ESClient == nil {
				return fmt.Errorf("elasticsearch is not available")
			}

			// Delete existing index if --clean flag is set
			if clean {
				l.Info("Cleaning existing index before reindex...")
				if err := cmdCtx.ESClient.DeleteProductIndices(ctx); err != nil {
					return fmt.Errorf("failed to delete index: %w", err)
				}
			}

			// Ensure index exists
			if err := cmdCtx.ESClient.EnsureProductIndices(ctx); err != nil {
				return fmt.Errorf("failed to initialize index: %w", err)
			}

			productRepo := repository.NewProductRepository(cmdCtx.GetMongoDatabase())

			totalSynced := 0
			cursor := ""

			for {
				products, nextCursor, err := productRepo.ListFeed(ctx, cursor, batchSize)
				if err != nil {
					return fmt.Errorf("failed to fetch products: %w", err)
				}

				if len(products) == 0 {
					break
				}

				docs := make([]*es.ProductDocument, 0, len(products))
				for i := range products {
					if products[i].Status == domain.StatusActive {
						docs = append(docs, buildProductDocument(&products[i]))
					}
				}

				if len(docs) > 0 {
					if err := cmdCtx.ESClient.BulkIndexProducts(ctx, docs); err != nil {
						l.Error("Failed to bulk index batch", err)
					} else {
						totalSynced += len(docs)
					}
				}

				l.Info("Batch indexed",
					logging.Int("batch_size", len(products)),
					logging.Int("total_synced", totalSynced),
				)

				if nextCursor == "" {
					break
				}
				cursor = nextCursor
			}

			l.Info(fmt.Sprintf("Reindex completed: %d products indexed", totalSynced),
				logging.Int("totalSynced", totalSynced),
			)
			return nil
		},
	}

	cmd.Flags().IntVar(&batchSize, "batch-size", 100, "Number of products per batch")
	cmd.Flags().BoolVar(&clean, "clean", false, "Delete existing index before reindexing")
	return cmd
}

// es health - Check Elasticsearch health
func newESHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check Elasticsearch connection health",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			l := logger.FromContext(ctx)

			cmdCtx, err := NewCommandContext()
			if err != nil {
				return fmt.Errorf("failed to create command context: %w", err)
			}
			defer cmdCtx.Close()

			if cmdCtx.ESClient == nil {
				return fmt.Errorf("elasticsearch is not available")
			}

			if err := cmdCtx.ESClient.Ping(); err != nil {
				return fmt.Errorf("elasticsearch health check failed: %w", err)
			}

			l.Info("Elasticsearch is healthy")
			return nil
		},
	}
}

func buildProductDocument(p *domain.Product) *es.ProductDocument {
	var categoryID, categoryName string
	if p.Category != nil {
		categoryID = p.Category.ID.Hex()
		categoryName = p.Category.Name
	}

	images := make([]string, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, img.URL)
	}

	return &es.ProductDocument{
		ID:           p.ID.Hex(),
		Slug:         p.Slug,
		Title:        p.Title,
		Description:  p.Description,
		Price:        p.Price,
		Brand:        p.Brand,
		Condition:    p.Condition,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		Status:       string(p.Status),
		SellerID:     p.SellerID.Hex(),
		Images:       images,
		Hashtags:     p.Hashtags,
		IsVerified:   p.IsVerified,
		IsFeatured:   p.IsFeatured,
		IsSelect:     p.IsSelect,
		PublishedAt:  p.PublishedAt,
		CreatedAt:    p.CreatedAt,
	}
}
