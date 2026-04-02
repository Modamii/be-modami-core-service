package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// EnsureProductIndices creates the language detection ingest pipeline and all
// language-specific product indices (base + one per SupportedLanguages).
//
// Pipeline behaviour (requires the ingest-langdetect ES plugin):
//  1. Detects language from the `title` field → writes to `language` field.
//  2. Reroutes the document to `{base}_{language}` via Painless script.
//     Falls back to the base index when language detection fails.
func (c *Client) EnsureProductIndices(ctx context.Context) error {
	if err := c.ensureLangDetectPipeline(ctx); err != nil {
		return fmt.Errorf("ensure langdetect pipeline: %w", err)
	}

	// Base (fallback) index — no language suffix
	if err := c.createProductIndex(ctx, c.Index, ""); err != nil {
		return fmt.Errorf("create base product index: %w", err)
	}

	// Language-specific indices
	for _, lang := range SupportedLanguages {
		name := ProductIndexNameForLang(c.Index, lang)
		if err := c.createProductIndex(ctx, name, lang); err != nil {
			return fmt.Errorf("create product index %s: %w", name, err)
		}
	}
	return nil
}

// DeleteProductIndices deletes the pipeline and all product indices.
func (c *Client) DeleteProductIndices(ctx context.Context) error {
	indices := []string{c.Index}
	for _, lang := range SupportedLanguages {
		indices = append(indices, ProductIndexNameForLang(c.Index, lang))
	}

	res, err := c.ES.Indices.Delete(
		indices,
		c.ES.Indices.Delete.WithContext(ctx),
		c.ES.Indices.Delete.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return fmt.Errorf("delete product indices: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("delete product indices status: %s", res.Status())
	}

	// Best-effort pipeline deletion
	pRes, err := c.ES.Ingest.DeletePipeline(
		PipelineProductLangIdent,
		c.ES.Ingest.DeletePipeline.WithContext(ctx),
	)
	if err == nil {
		pRes.Body.Close()
	}
	return nil
}

// ensureLangDetectPipeline creates the ingest pipeline that:
//  1. Detects language from the product `title` field using the ingest-langdetect plugin.
//  2. Reroutes the document to the matching language index via a Painless script.
func (c *Client) ensureLangDetectPipeline(ctx context.Context) error {
	// Check if pipeline already exists (404 = not found, proceed to create)
	res, err := c.ES.Ingest.GetPipeline(
		c.ES.Ingest.GetPipeline.WithPipelineID(PipelineProductLangIdent),
		c.ES.Ingest.GetPipeline.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("check pipeline: %w", err)
	}
	res.Body.Close()
	if res.StatusCode == 200 {
		return nil // already exists
	}

	pipeline := map[string]any{
		"description": "Detect language from product title and route to language-specific index",
		"processors": []any{
			// Step 1: detect language via the ingest-langdetect plugin
			map[string]any{
				"langdetect": map[string]any{
					"field":          "title",
					"target_field":   "language",
					"ignore_missing": true,
					"on_failure": []any{
						map[string]any{
							"set": map[string]any{
								"field": "language",
								"value": "",
							},
						},
					},
				},
			},
			// Step 2: reroute to language-specific index when language is known
			map[string]any{
				"script": map[string]any{
					"lang": "painless",
					"source": `
						String lang = ctx.containsKey('language') ? ctx['language'] : '';
						if (lang != null && !lang.isEmpty()) {
							ctx['_index'] = ctx['_index'] + '_' + lang;
						}
					`,
				},
			},
		},
	}

	body, _ := json.Marshal(pipeline)
	putRes, err := c.ES.Ingest.PutPipeline(
		PipelineProductLangIdent,
		bytes.NewReader(body),
		c.ES.Ingest.PutPipeline.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create langdetect pipeline: %w", err)
	}
	defer putRes.Body.Close()
	if putRes.IsError() {
		return fmt.Errorf("create langdetect pipeline status: %s", putRes.Status())
	}
	return nil
}

// createProductIndex creates one product index with language-specific analyzer settings
// and shared field mappings. Skips creation if the index already exists.
func (c *Client) createProductIndex(ctx context.Context, indexName, lang string) error {
	res, err := c.ES.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("check index %s: %w", indexName, err)
	}
	res.Body.Close()
	if res.StatusCode == 200 {
		return nil // already exists
	}

	body := map[string]any{
		"settings": languageAnalyzerSettings(lang),
		"mappings": ProductFieldMappings(),
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal index body: %w", err)
	}

	createRes, err := c.ES.Indices.Create(
		indexName,
		c.ES.Indices.Create.WithBody(strings.NewReader(string(encoded))),
		c.ES.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create index %s: %w", indexName, err)
	}
	defer createRes.Body.Close()
	if createRes.IsError() {
		return fmt.Errorf("create index %s status: %s", indexName, createRes.Status())
	}
	return nil
}
