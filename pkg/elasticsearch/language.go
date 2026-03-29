package elasticsearch

import "fmt"

// Supported languages for multi-language indexing.
var SupportedLanguages = []string{"vi", "en", "ja", "zh", "ko", "ru", "es"}

const (
	PipelineLangIdent = "pipe_lang_ident_article"
)

// ArticleIndexName returns the default articles index name: {namespace}_articles
func ArticleIndexName(namespace string) string {
	return namespace + "_articles"
}

// ArticleIndexNameForLang returns the language-specific index: {namespace}_articles_{lang}
func ArticleIndexNameForLang(namespace, lang string) string {
	if lang == "" {
		return ArticleIndexName(namespace)
	}
	return fmt.Sprintf("%s_articles_%s", namespace, lang)
}

// ArticleIndexWildcard returns the wildcard pattern to search all language indices: {namespace}_articles*
func ArticleIndexWildcard(namespace string) string {
	return namespace + "_articles*"
}

// langMapping holds per-language analyzer settings and field mappings.
type langMapping struct {
	Lang     string
	Settings map[string]any
}

// GetLanguageMappings returns analyzer settings for each language.
func GetLanguageMappings() []langMapping {
	return []langMapping{
		{
			Lang: "", // default
			Settings: map[string]any{
				"analysis": map[string]any{
					"analyzer": map[string]any{
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "icu_tokenizer",
							"filter":      []string{"icu_normalizer", "lowercase", "icu_folding"},
							"char_filter": []string{"html_strip"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "icu_tokenizer",
							"filter":      []string{"icu_normalizer", "lowercase"},
							"char_filter": []string{"html_strip"},
						},
					},
				},
			},
		},
		{
			Lang: "vi",
			Settings: map[string]any{
				"analysis": map[string]any{
					"analyzer": map[string]any{
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "uax_url_email",
							"filter":      []string{"icu_normalizer", "lowercase", "asciifolding"},
							"char_filter": []string{"html_strip"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "uax_url_email",
							"filter":      []string{"icu_normalizer", "lowercase"},
							"char_filter": []string{"html_strip"},
						},
					},
				},
			},
		},
		{
			Lang: "en",
			Settings: map[string]any{
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
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "uax_url_email",
							"filter":      []string{"english_possessive_stemmer", "lowercase", "english_stop", "english_stemmer"},
							"char_filter": []string{"html_strip"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "uax_url_email",
							"filter":      []string{"english_possessive_stemmer", "lowercase", "english_stop", "english_stemmer"},
							"char_filter": []string{"html_strip"},
						},
					},
				},
			},
		},
		{
			Lang: "ja",
			Settings: map[string]any{
				"analysis": map[string]any{
					"analyzer": map[string]any{
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "kuromoji_tokenizer",
							"filter":      []string{"kuromoji_baseform", "kuromoji_part_of_speech", "cjk_width", "ja_stop", "kuromoji_stemmer", "lowercase"},
							"char_filter": []string{"html_strip", "icu_normalizer"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "kuromoji_tokenizer",
							"filter":      []string{"kuromoji_baseform", "kuromoji_part_of_speech", "cjk_width", "ja_stop", "kuromoji_stemmer", "lowercase"},
							"char_filter": []string{"html_strip", "icu_normalizer"},
						},
					},
				},
			},
		},
		{
			Lang: "zh",
			Settings: map[string]any{
				"analysis": map[string]any{
					"analyzer": map[string]any{
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "smartcn_tokenizer",
							"filter":      []string{"lowercase", "smartcn_stop", "porter_stem"},
							"char_filter": []string{"html_strip"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "smartcn_tokenizer",
							"filter":      []string{"lowercase", "smartcn_stop"},
							"char_filter": []string{"html_strip"},
						},
					},
				},
			},
		},
		{
			Lang: "ko",
			Settings: map[string]any{
				"analysis": map[string]any{
					"analyzer": map[string]any{
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "nori_tokenizer",
							"filter":      []string{"lowercase"},
							"char_filter": []string{"html_strip", "icu_normalizer"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "nori_tokenizer",
							"filter":      []string{"lowercase"},
							"char_filter": []string{"html_strip", "icu_normalizer"},
						},
					},
				},
			},
		},
		{
			Lang: "ru",
			Settings: map[string]any{
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
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "standard",
							"filter":      []string{"lowercase", "russian_stop", "russian_stemmer"},
							"char_filter": []string{"html_strip"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "standard",
							"filter":      []string{"lowercase", "russian_stop", "russian_stemmer"},
							"char_filter": []string{"html_strip"},
						},
					},
				},
			},
		},
		{
			Lang: "es",
			Settings: map[string]any{
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
						"article_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "standard",
							"filter":      []string{"lowercase", "spanish_stop", "spanish_stemmer", "asciifolding"},
							"char_filter": []string{"html_strip"},
						},
						"article_no_ascii": map[string]any{
							"type":        "custom",
							"tokenizer":   "standard",
							"filter":      []string{"lowercase", "spanish_stop", "spanish_stemmer"},
							"char_filter": []string{"html_strip"},
						},
					},
				},
			},
		},
	}
}

// ArticleFieldMappings returns the common field mappings for article indices.
func ArticleFieldMappings() map[string]any {
	return map[string]any{
		"dynamic": false,
		"properties": map[string]any{
			"id":   map[string]any{"type": "keyword"},
			"slug": map[string]any{"type": "keyword"},
			"title": map[string]any{
				"type":     "text",
				"analyzer": "article_no_ascii",
				"fields": map[string]any{
					"ascii":   map[string]any{"type": "text", "analyzer": "article_ascii"},
					"keyword": map[string]any{"type": "keyword"},
				},
			},
			"description": map[string]any{
				"type":     "text",
				"analyzer": "article_no_ascii",
				"fields": map[string]any{
					"ascii": map[string]any{"type": "text", "analyzer": "article_ascii"},
				},
			},
			"content": map[string]any{
				"type":     "text",
				"analyzer": "article_no_ascii",
				"fields": map[string]any{
					"ascii": map[string]any{"type": "text", "analyzer": "article_ascii"},
				},
			},
			"authorId":     map[string]any{"type": "keyword"},
			"authorName":   map[string]any{"type": "text", "analyzer": "article_ascii"},
			"categoryId":   map[string]any{"type": "keyword"},
			"categoryName": map[string]any{"type": "text", "analyzer": "article_ascii"},
			"categorySlug": map[string]any{"type": "keyword"},
			"tags": map[string]any{
				"type": "nested",
				"properties": map[string]any{
					"id":   map[string]any{"type": "keyword"},
					"name": map[string]any{"type": "text", "analyzer": "article_ascii"},
				},
			},
			"image":       map[string]any{"type": "keyword", "index": false},
			"status":      map[string]any{"type": "keyword"},
			"isPremium":   map[string]any{"type": "boolean"},
			"isFeatured":  map[string]any{"type": "boolean"},
			"readTime":    map[string]any{"type": "keyword"},
			"publishedAt": map[string]any{"type": "date"},
			"createdAt":   map[string]any{"type": "date"},
			"updatedAt":   map[string]any{"type": "date"},
		},
	}
}
