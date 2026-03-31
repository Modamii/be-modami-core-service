package pagination

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const DefaultLimit = 20
const MaxLimit = 100

type CursorParams struct {
	Cursor string
	Limit  int
}

type CursorData struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"ca"`
}

type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

func ParseCursor(r *http.Request) CursorParams {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > MaxLimit {
		limit = DefaultLimit
	}
	return CursorParams{
		Cursor: r.URL.Query().Get("cursor"),
		Limit:  limit,
	}
}

func DecodeCursor(cursor string) (*CursorData, error) {
	if cursor == "" {
		return nil, nil
	}
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var data CursorData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func EncodeCursor(id string, createdAt time.Time) string {
	data := CursorData{ID: id, CreatedAt: createdAt}
	b, _ := json.Marshal(data)
	return base64.URLEncoding.EncodeToString(b)
}

func CursorFilter(cursor string, sortField string) (bson.D, error) {
	data, err := DecodeCursor(cursor)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return bson.D{}, nil
	}

	oid, err := bson.ObjectIDFromHex(data.ID)
	if err != nil {
		return nil, err
	}

	return bson.D{
		{Key: "$or", Value: bson.A{
			bson.D{
				{Key: sortField, Value: bson.D{{Key: "$lt", Value: data.CreatedAt}}},
			},
			bson.D{
				{Key: sortField, Value: data.CreatedAt},
				{Key: "_id", Value: bson.D{{Key: "$lt", Value: oid}}},
			},
		}},
	}, nil
}
