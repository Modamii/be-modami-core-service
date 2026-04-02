package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Report struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ReporterID string        `bson:"reporter_id" json:"reporter_id"`
	TargetType string        `bson:"target_type" json:"target_type"`
	TargetID   bson.ObjectID `bson:"target_id" json:"target_id"`
	Reason     string        `bson:"reason" json:"reason"`
	Detail     string        `bson:"detail,omitempty" json:"detail,omitempty"`
	Status     string        `bson:"status" json:"status"`
	ResolvedBy *string       `bson:"resolved_by,omitempty" json:"resolved_by,omitempty"`
	ResolvedAt *time.Time    `bson:"resolved_at,omitempty" json:"resolved_at,omitempty"`
	Resolution string        `bson:"resolution,omitempty" json:"resolution,omitempty"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
}
