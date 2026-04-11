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
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Database     int           `mapstructure:"database"`
	Pass         string        `mapstructure:"pass"`
	UserName     string        `mapstructure:"user_name"`
	PoolSize     int           `mapstructure:"pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	BrokerList    string `mapstructure:"broker_list"`
	Enable        bool   `mapstructure:"enable"`
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
	Enable   bool   `mapstructure:"enable"`
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

	v.SetDefault("app.name", "Modami core")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.port", 8080)
	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.swagger_host", "localhost:8087")
	v.SetDefault("app.shutdown_timeout", "30s")
	v.SetDefault("app.read_timeout", "30s")
	v.SetDefault("app.write_timeout", "30s")
	v.SetDefault("app.idle_timeout", "120s")

	v.SetDefault("mongodb.uri", "mongodb://localhost:27017")
	v.SetDefault("mongodb.database", "modami")
	v.SetDefault("observability.log_level", "info")
	v.SetDefault("observability.environment", "development")
	v.SetDefault("kafka.env", "development")
	v.SetDefault("app.allow_credentials", true)
	v.SetDefault("app.allowed_origins", []string{
		"http://localhost:5173",
		"http://localhost:3000",
		"http://localhost:8080",
		"http://localhost:8081",
	})

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
