package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Mongo         MongoConfig         `mapstructure:"mongodb"`
	Keycloak      KeycloakConfig      `mapstructure:"keycloak"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Observability ObservabilityConfig `mapstructure:"observability"`
}

// AppConfig holds process / HTTP server settings (merged from YAML `app:`).
type AppConfig struct {
	Name             string   `mapstructure:"name"`
	Version          string   `mapstructure:"version"`
	Environment      string   `mapstructure:"environment"`
	Debug            bool     `mapstructure:"debug"`
	Port             int      `mapstructure:"port"`
	Host             string   `mapstructure:"host"`
	SwaggerHost      string   `mapstructure:"swagger_host"`
	ShutdownTimeout  string   `mapstructure:"shutdown_timeout"`
	ReadTimeout      string   `mapstructure:"read_timeout"`
	WriteTimeout     string   `mapstructure:"write_timeout"`
	IdleTimeout      string   `mapstructure:"idle_timeout"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
}

// ListenAddr returns host:port for http.Server (defaults host 0.0.0.0, port 8080).
func (a AppConfig) ListenAddr() string {
	host := strings.TrimSpace(a.Host)
	if host == "" {
		host = "0.0.0.0"
	}
	port := a.Port
	if port <= 0 {
		port = 8080
	}
	return fmt.Sprintf("%s:%d", host, port)
}

// PortString returns the port as a decimal string (for logs).
func (a AppConfig) PortString() string {
	p := a.Port
	if p <= 0 {
		p = 8080
	}
	return fmt.Sprintf("%d", p)
}

func (a AppConfig) GetShutdownTimeout() time.Duration {
	d, err := time.ParseDuration(a.ShutdownTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetReadTimeout() time.Duration {
	d, err := time.ParseDuration(a.ReadTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetWriteTimeout() time.Duration {
	d, err := time.ParseDuration(a.WriteTimeout)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func (a AppConfig) GetIdleTimeout() time.Duration {
	d, err := time.ParseDuration(a.IdleTimeout)
	if err != nil {
		return 120 * time.Second
	}
	return d
}

type MongoConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}

type KeycloakConfig struct {
	JWKSUrl string `mapstructure:"jwks_url"`
}

type RedisConfig struct {
	Host              string         `mapstructure:"host"`
	Port              int            `mapstructure:"port"`
	Database          int            `mapstructure:"database"`
	RateLimitDatabase int            `mapstructure:"rate_limit_database"`
	TTL               string         `mapstructure:"ttl"`
	PoolSize          int            `mapstructure:"pool_size"`
	Pass              string         `mapstructure:"pass"`
	UserName          string         `mapstructure:"user_name"`
	WriteTimeout      string         `mapstructure:"write_timeout"`
	ReadTimeout       string         `mapstructure:"read_timeout"`
	DialTimeout       string         `mapstructure:"dial_timeout"`
	TLSConfig         RedisTLSConfig `mapstructure:"tls_config"`
}

type RedisTLSConfig struct {
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func (r RedisConfig) GetDialTimeout() time.Duration {
	d, err := time.ParseDuration(r.DialTimeout)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

func (r RedisConfig) GetReadTimeout() time.Duration {
	d, err := time.ParseDuration(r.ReadTimeout)
	if err != nil {
		return 3 * time.Second
	}
	return d
}

func (r RedisConfig) GetWriteTimeout() time.Duration {
	d, err := time.ParseDuration(r.WriteTimeout)
	if err != nil {
		return 3 * time.Second
	}
	return d
}

type KafkaConfig struct {
	BrokerList    string `mapstructure:"broker_list"`
	Env           string `mapstructure:"env"`
	ClientID      string `mapstructure:"client_id"`
	ConsumerGroup string `mapstructure:"consumer_group"`
}

func (k KafkaConfig) Brokers() []string {
	if k.BrokerList == "" {
		return nil
	}
	return strings.Split(k.BrokerList, ",")
}

type ElasticsearchConfig struct {
	URL      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Index    string `mapstructure:"index"`
}

type ObservabilityConfig struct {
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`
	Environment    string `mapstructure:"environment"`
	LogLevel       string `mapstructure:"log_level"`
	OTLPEndpoint   string `mapstructure:"otlp_endpoint"`
	OTLPInsecure   bool   `mapstructure:"otlp_insecure"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
