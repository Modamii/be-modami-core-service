package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pkges "gitlab.com/lifegoeson-libs/pkg-gokit/elasticsearch"
)

const PipelineProductLangIdent = "pipe_lang_ident_product"

// ProductDocument is the Elasticsearch document for a product.
type ProductDocument struct {
	ID           string     `json:"id"`
	Slug         string     `json:"slug"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Price        int64      `json:"price"`
	Brand        string     `json:"brand,omitempty"`
	Condition    string     `json:"condition"`
	CategoryID   string     `json:"category_id,omitempty"`
	CategoryName string     `json:"category_name,omitempty"`
	Status       string     `json:"status"`
	SellerID     string     `json:"seller_id"`
	Images       []string   `json:"images,omitempty"`
	Hashtags     []string   `json:"hashtags,omitempty"`
	IsVerified   bool       `json:"is_verified"`
	IsFeatured   bool       `json:"is_featured"`
	IsSelect     bool       `json:"is_select"`
	Language     string     `json:"language,omitempty"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// Ping checks connectivity to Elasticsearch.
func Ping(c *pkges.Client) error {
	res, err := c.ES.Ping()
	if err != nil {
		return fmt.Errorf("elasticsearch ping: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch ping returned: %s", res.Status())
	}
	return nil
}

// IndexProduct indexes a single product through the language detection pipeline.
func IndexProduct(ctx context.Context, c *pkges.Client, doc *ProductDocument) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal product doc: %w", err)
	}
	res, err := c.ES.Index(
		c.Index,
		bytes.NewReader(body),
		c.ES.Index.WithDocumentID(doc.ID),
		c.ES.Index.WithPipeline(PipelineProductLangIdent),
		c.ES.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("es index product: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es index error: %s", res.Status())
	}
	return nil
}

// DeleteProduct removes a product from the base index and all language indices.
func DeleteProduct(ctx context.Context, c *pkges.Client, id string) error {
	if err := deleteFromIndex(ctx, c, c.Index, id); err != nil {
		return err
	}
	for _, lang := range pkges.SupportedLanguages {
		_ = deleteFromIndex(ctx, c, pkges.IndexNameForLang(c.Index, lang), id)
	}
	return nil
}

// BulkIndexProducts indexes a batch of product documents using the bulk API.
func BulkIndexProducts(ctx context.Context, c *pkges.Client, docs []*ProductDocument) error {
	if len(docs) == 0 {
		return nil
	}
	var buf strings.Builder
	for _, doc := range docs {
		meta := fmt.Sprintf(`{"index":{"_index":%q,"_id":%q,"pipeline":%q}}`,
			c.Index, doc.ID, PipelineProductLangIdent)
		buf.WriteString(meta)
		buf.WriteByte('\n')
		body, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("marshal product %s: %w", doc.ID, err)
		}
		buf.Write(body)
		buf.WriteByte('\n')
	}
	res, err := c.ES.Bulk(
		strings.NewReader(buf.String()),
		c.ES.Bulk.WithContext(ctx),
		c.ES.Bulk.WithIndex(c.Index),
	)
	if err != nil {
		return fmt.Errorf("bulk index products: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("bulk index error: %s", res.Status())
	}
	return nil
}

func deleteFromIndex(ctx context.Context, c *pkges.Client, index, id string) error {
	res, err := c.ES.Delete(index, id, c.ES.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("es delete product: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("es delete error: %s", res.Status())
	}
	return nil
}
