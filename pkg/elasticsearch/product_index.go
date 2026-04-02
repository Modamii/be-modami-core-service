package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

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

// IndexProduct indexes a single product document through the language detection pipeline.
// The pipeline (pipe_lang_ident_product) detects the language from the title field and
// automatically routes the document to the appropriate language-specific index.
func (c *Client) IndexProduct(ctx context.Context, doc *ProductDocument) error {
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

// DeleteProduct removes a product from all language indices.
func (c *Client) DeleteProduct(ctx context.Context, id string) error {
	// Delete from base index
	if err := c.deleteFromIndex(ctx, c.Index, id); err != nil {
		return err
	}
	// Delete from all language-specific indices (ignore 404)
	for _, lang := range SupportedLanguages {
		_ = c.deleteFromIndex(ctx, ProductIndexNameForLang(c.Index, lang), id)
	}
	return nil
}

func (c *Client) deleteFromIndex(ctx context.Context, index, id string) error {
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

// BulkIndexProducts indexes a batch of product documents using the bulk API.
// Each document is sent through the language detection pipeline.
func (c *Client) BulkIndexProducts(ctx context.Context, docs []*ProductDocument) error {
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
