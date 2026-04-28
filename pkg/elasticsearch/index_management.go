package elasticsearch

import (
	"context"
	"fmt"

	pkges "gitlab.com/lifegoeson-libs/pkg-gokit/elasticsearch"
)


// productMappings defines the Elasticsearch field mappings for product documents.
var productMappings = map[string]any{
	"properties": map[string]any{
		"id":   map[string]any{"type": "keyword"},
		"slug": map[string]any{"type": "keyword"},
		"title": map[string]any{
			"type":     "text",
			"analyzer": "product_ascii",
			"fields": map[string]any{
				"no_ascii": map[string]any{"type": "text", "analyzer": "product_no_ascii"},
				"raw":      map[string]any{"type": "keyword"},
			},
		},
		"description": map[string]any{
			"type":     "text",
			"analyzer": "product_ascii",
		},
		"brand":         map[string]any{"type": "keyword"},
		"condition":     map[string]any{"type": "keyword"},
		"category_id":   map[string]any{"type": "keyword"},
		"category_name": map[string]any{"type": "keyword"},
		"status":        map[string]any{"type": "keyword"},
		"seller_id":     map[string]any{"type": "keyword"},
		"hashtags":      map[string]any{"type": "keyword"},
		"price":         map[string]any{"type": "long"},
		"is_verified":   map[string]any{"type": "boolean"},
		"is_featured":   map[string]any{"type": "boolean"},
		"is_select":     map[string]any{"type": "boolean"},
		"language":      map[string]any{"type": "keyword"},
		"images":        map[string]any{"type": "keyword"},
		"published_at":  map[string]any{"type": "date"},
		"created_at":    map[string]any{"type": "date"},
	},
}

// productLangPipeline is the ingest pipeline definition for language-aware routing.
var productLangPipeline = map[string]any{
	"description": "Language detection and routing pipeline for products",
	"processors": []map[string]any{
		{
			"inference": map[string]any{
				"model_id":     "lang_ident_model_1",
				"field_map":    map[string]any{"title": "text_field"},
				"target_field": "_ingest._value",
				"on_failure": []map[string]any{
					{"set": map[string]any{"field": "language", "value": "vi"}},
				},
			},
		},
		{
			"set": map[string]any{
				"field": "language",
				"value": "{{_ingest._value.predicted_value}}",
			},
		},
		{
			"reroute": map[string]any{
				"dataset": "{{language}}",
			},
		},
	},
}

// EnsureProductIndices creates the language detection pipeline and all language-specific
// product indices with proper analyzers and mappings.
func EnsureProductIndices(ctx context.Context, c *pkges.Client) error {
	if err := c.EnsurePipeline(ctx, PipelineProductLangIdent, productLangPipeline); err != nil {
		return fmt.Errorf("ensure product lang pipeline: %w", err)
	}
	for _, lang := range pkges.SupportedLanguages {
		indexName := pkges.IndexNameForLang(c.Index, lang)
		settings := pkges.LanguageAnalyzerSettings(lang)
		if err := c.EnsureIndex(ctx, indexName, settings, productMappings); err != nil {
			return fmt.Errorf("ensure product index %s: %w", indexName, err)
		}
	}
	return nil
}

// DeleteProductIndices removes all product language-specific indices.
func DeleteProductIndices(ctx context.Context, c *pkges.Client) error {
	names := make([]string, 0, len(pkges.SupportedLanguages)+1)
	names = append(names, c.Index)
	for _, lang := range pkges.SupportedLanguages {
		names = append(names, pkges.IndexNameForLang(c.Index, lang))
	}
	return c.DeleteIndices(ctx, names...)
}
