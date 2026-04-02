package elasticsearch

import "fmt"

// SupportedLanguages lists all language codes with dedicated product indices.
var SupportedLanguages = []string{"vi", "en", "ja", "zh", "ko", "ru", "es"}

// PipelineProductLangIdent is the ingest pipeline name for language detection.
const PipelineProductLangIdent = "pipe_lang_ident_product"

// ProductIndexNameForLang returns the language-specific index name.
// An empty lang returns the base index (used as the fallback / write target).
func ProductIndexNameForLang(base, lang string) string {
	if lang == "" {
		return base
	}
	return fmt.Sprintf("%s_%s", base, lang)
}

// ProductIndexWildcard returns a wildcard pattern matching all language indices.
func ProductIndexWildcard(base string) string {
	return base + "*"
}

// ProductFieldMappings returns the common field mappings applied to every
// language-specific product index.
func ProductFieldMappings() map[string]any {
	return map[string]any{
		"dynamic": false,
		"properties": map[string]any{
			"id":   map[string]any{"type": "keyword"},
			"slug": map[string]any{"type": "keyword"},
			"title": map[string]any{
				"type":     "text",
				"analyzer": "product_no_ascii",
				"fields": map[string]any{
					"ascii":   map[string]any{"type": "text", "analyzer": "product_ascii"},
					"keyword": map[string]any{"type": "keyword"},
				},
			},
			"description": map[string]any{
				"type":     "text",
				"analyzer": "product_no_ascii",
				"fields": map[string]any{
					"ascii": map[string]any{"type": "text", "analyzer": "product_ascii"},
				},
			},
			"brand":         map[string]any{"type": "keyword"},
			"condition":     map[string]any{"type": "keyword"},
			"price":         map[string]any{"type": "long"},
			"category_id":   map[string]any{"type": "keyword"},
			"category_name": map[string]any{"type": "text", "analyzer": "product_ascii"},
			"status":        map[string]any{"type": "keyword"},
			"seller_id":     map[string]any{"type": "keyword"},
			"images":        map[string]any{"type": "keyword", "index": false},
			"hashtags":      map[string]any{"type": "keyword"},
			"is_verified":   map[string]any{"type": "boolean"},
			"is_featured":   map[string]any{"type": "boolean"},
			"is_select":     map[string]any{"type": "boolean"},
			"language":      map[string]any{"type": "keyword"},
			"published_at":  map[string]any{"type": "date"},
			"created_at":    map[string]any{"type": "date"},
		},
	}
}

// languageAnalyzerSettings returns the analysis settings for each supported language.
// The analyzers are named `product_ascii` and `product_no_ascii` so every index
// can use the same field mappings regardless of language.
func languageAnalyzerSettings(lang string) map[string]any {
	switch lang {
	case "vi":
		return map[string]any{
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "uax_url_email",
						"filter":      []string{"icu_normalizer", "lowercase", "asciifolding"},
						"char_filter": []string{"html_strip"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "uax_url_email",
						"filter":      []string{"icu_normalizer", "lowercase"},
						"char_filter": []string{"html_strip"},
					},
				},
			},
		}
	case "en":
		return map[string]any{
			"analysis": map[string]any{
				"filter": map[string]any{
					"english_stop": map[string]any{
						"type":      "stop",
						"stopwords": "_english_",
					},
					"english_stemmer": map[string]any{
						"type":     "stemmer",
						"language": "english",
					},
					"english_possessive_stemmer": map[string]any{
						"type":     "stemmer",
						"language": "possessive_english",
					},
				},
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "uax_url_email",
						"filter":      []string{"english_possessive_stemmer", "lowercase", "english_stop", "english_stemmer"},
						"char_filter": []string{"html_strip"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "uax_url_email",
						"filter":      []string{"english_possessive_stemmer", "lowercase", "english_stop", "english_stemmer"},
						"char_filter": []string{"html_strip"},
					},
				},
			},
		}
	case "ja":
		return map[string]any{
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "kuromoji_tokenizer",
						"filter":      []string{"kuromoji_baseform", "kuromoji_part_of_speech", "cjk_width", "ja_stop", "kuromoji_stemmer", "lowercase"},
						"char_filter": []string{"html_strip", "icu_normalizer"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "kuromoji_tokenizer",
						"filter":      []string{"kuromoji_baseform", "kuromoji_part_of_speech", "cjk_width", "ja_stop", "kuromoji_stemmer", "lowercase"},
						"char_filter": []string{"html_strip", "icu_normalizer"},
					},
				},
			},
		}
	case "zh":
		return map[string]any{
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "smartcn_tokenizer",
						"filter":      []string{"lowercase", "smartcn_stop", "porter_stem"},
						"char_filter": []string{"html_strip"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "smartcn_tokenizer",
						"filter":      []string{"lowercase", "smartcn_stop"},
						"char_filter": []string{"html_strip"},
					},
				},
			},
		}
	case "ko":
		return map[string]any{
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "nori_tokenizer",
						"filter":      []string{"lowercase"},
						"char_filter": []string{"html_strip", "icu_normalizer"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "nori_tokenizer",
						"filter":      []string{"lowercase"},
						"char_filter": []string{"html_strip", "icu_normalizer"},
					},
				},
			},
		}
	case "ru":
		return map[string]any{
			"analysis": map[string]any{
				"filter": map[string]any{
					"russian_stop": map[string]any{
						"type":      "stop",
						"stopwords": "_russian_",
					},
					"russian_stemmer": map[string]any{
						"type":     "stemmer",
						"language": "russian",
					},
				},
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "standard",
						"filter":      []string{"lowercase", "russian_stop", "russian_stemmer"},
						"char_filter": []string{"html_strip"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "standard",
						"filter":      []string{"lowercase", "russian_stop", "russian_stemmer"},
						"char_filter": []string{"html_strip"},
					},
				},
			},
		}
	case "es":
		return map[string]any{
			"analysis": map[string]any{
				"filter": map[string]any{
					"spanish_stop": map[string]any{
						"type":      "stop",
						"stopwords": "_spanish_",
					},
					"spanish_stemmer": map[string]any{
						"type":     "stemmer",
						"language": "light_spanish",
					},
				},
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "standard",
						"filter":      []string{"lowercase", "spanish_stop", "spanish_stemmer", "asciifolding"},
						"char_filter": []string{"html_strip"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "standard",
						"filter":      []string{"lowercase", "spanish_stop", "spanish_stemmer"},
						"char_filter": []string{"html_strip"},
					},
				},
			},
		}
	default:
		// ICU-based default — works for any script
		return map[string]any{
			"analysis": map[string]any{
				"analyzer": map[string]any{
					"product_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "icu_tokenizer",
						"filter":      []string{"icu_normalizer", "lowercase", "icu_folding"},
						"char_filter": []string{"html_strip"},
					},
					"product_no_ascii": map[string]any{
						"type":        "custom",
						"tokenizer":   "icu_tokenizer",
						"filter":      []string{"icu_normalizer", "lowercase"},
						"char_filter": []string{"html_strip"},
					},
				},
			},
		}
	}
}
