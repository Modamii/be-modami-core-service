package elasticsearch

import (
	"fmt"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
)

type Config struct {
	URL      string
	Username string
	Password string
	Index    string
}

type Client struct {
	ES    *elasticsearch.Client
	Index string
}

func NewClient(cfg *Config) (*Client, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.URL},
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}

	if cfg.Username != "" && cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &Client{
		ES:    es,
		Index: cfg.Index,
	}, nil
}

func (c *Client) Ping() error {
	res, err := c.ES.Ping()
	if err != nil {
		return fmt.Errorf("elasticsearch ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch ping returned status: %s", res.Status())
	}
	return nil
}
