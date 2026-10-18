package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Environment string

const (
	EnvTest       Environment = "test"
	EnvLocal      Environment = "local"
	EnvHomolog    Environment = "homolog"
	EnvSandbox    Environment = "sandbox"
	EnvProduction Environment = "production"
)

type Config struct {
	Environment Environment `envconfig:"ENVIRONMENT" default:"local"`
	Development bool        `envconfig:"DEVELOPMENT" default:"true"`

	App      App
	Server   Server
	Postgres Postgres
	Demo     Demo
	Auth     Auth
	OpenAI   OpenAI
	Reports  Reports
}

type App struct {
	Name                    string        `envconfig:"APP_NAME"                      default:"SquadFlow"`
	ID                      string        `envconfig:"APP_ID"                        default:"squadflow-api"`
	GracefulShutdownTimeout time.Duration `envconfig:"APP_GRACEFUL_SHUTDOWN_TIMEOUT" default:"10s"`
	Version                 string        `ignored:"true"`
}

type Server struct {
	Address           string        `envconfig:"SERVER_ADDRESS"             default:":8090"`
	ReadTimeout       time.Duration `envconfig:"SERVER_READ_TIMEOUT"        default:"15s"`
	ReadHeaderTimeout time.Duration `envconfig:"SERVER_READ_HEADER_TIMEOUT" default:"5s"`
	WriteTimeout      time.Duration `envconfig:"SERVER_WRITE_TIMEOUT"       default:"30s"`
	IdleTimeout       time.Duration `envconfig:"SERVER_IDLE_TIMEOUT"        default:"60s"`
}

type Postgres struct {
	// Em Docker, DATABASE_URL tem prioridade sobre os campos DB_*.
	DatabaseURL string `envconfig:"DATABASE_URL"`

	Host         string `envconfig:"DB_HOST"     default:"localhost"`
	Port         string `envconfig:"DB_PORT"     default:"5432"`
	User         string `envconfig:"DB_USER"     default:"postgres"`
	Password     string `envconfig:"DB_PASSWORD"`
	DatabaseName string `envconfig:"DB_NAME"     default:"squadflow"`
	SSLMode      string `envconfig:"DB_SSLMODE"  default:"disable"`
}

type Demo struct {
	Seed bool `envconfig:"DEMO_SEED" default:"false"`
}

type Auth struct {
	JWTSecret      string        `envconfig:"JWT_SECRET"`
	AccessTTL      time.Duration `envconfig:"ACCESS_TOKEN_TTL"  default:"15m"`
	RefreshTTL     time.Duration `envconfig:"REFRESH_TOKEN_TTL" default:"168h"`
	CookieSecure   bool          `envconfig:"COOKIE_SECURE"     default:"false"`
	AllowedOrigins string        `envconfig:"ALLOWED_ORIGINS"   default:"http://localhost:5173,http://127.0.0.1:5173,http://localhost:8080,http://127.0.0.1:8080"`
	AdminName      string        `envconfig:"ADMIN_NAME"        default:"Administrador"`
	AdminEmail     string        `envconfig:"ADMIN_EMAIL"`
	AdminPassword  string        `envconfig:"ADMIN_PASSWORD"`
}
type OpenAI struct {
	APIKey         string        `envconfig:"OPENAI_API_KEY"`
	Model          string        `envconfig:"OPENAI_MODEL"           default:"gpt-4o-mini"`
	Timeout        time.Duration `envconfig:"AI_REPORT_TIMEOUT"      default:"210s"`
	StoreResponses bool          `envconfig:"OPENAI_STORE_RESPONSES" default:"false"`
}

type Reports struct {
	Role        string `envconfig:"PROCESS_ROLE"              default:"all"`
	RedisURL    string `envconfig:"REDIS_URL"                 default:"redis://localhost:6379/0"`
	Stream      string `envconfig:"REPORT_QUEUE_STREAM"       default:"squadflow:reports"`
	Group       string `envconfig:"REPORT_QUEUE_GROUP"        default:"report-workers"`
	Concurrency int    `envconfig:"REPORT_WORKER_CONCURRENCY" default:"2"`
}

// ConnectionURL devolve a URL que deve ser passada para postgres.New.
func (p Postgres) ConnectionURL() string {
	if strings.TrimSpace(p.DatabaseURL) != "" {
		return p.DatabaseURL
	}

	credentials := url.User(p.User)
	if p.Password != "" {
		credentials = url.UserPassword(p.User, p.Password)
	}

	query := url.Values{}
	query.Set("sslmode", p.SSLMode)

	return (&url.URL{
		Scheme:   "postgres",
		User:     credentials,
		Host:     net.JoinHostPort(p.Host, p.Port),
		Path:     "/" + p.DatabaseName,
		RawQuery: query.Encode(),
	}).String()
}

func New(version string) (Config, error) {
	const operation = "Config.New"

	// Ausência de .env não é erro: containers podem receber variáveis
	// diretamente do ambiente. Load não sobrescreve variáveis já existentes.
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("%s: %w", operation, err)
	}

	switch cfg.Environment {
	case EnvTest, EnvLocal, EnvHomolog, EnvSandbox, EnvProduction:
	default:
		return Config{}, fmt.Errorf("%s: invalid ENVIRONMENT %q", operation, cfg.Environment)
	}

	if cfg.Environment == EnvProduction && cfg.Demo.Seed {
		return Config{}, fmt.Errorf("%s: DEMO_SEED cannot be enabled in production", operation)
	}

	if len(cfg.Auth.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("%s: JWT_SECRET must contain at least 32 bytes", operation)
	}

	if cfg.Auth.AccessTTL <= 0 || cfg.Auth.RefreshTTL <= cfg.Auth.AccessTTL {
		return Config{}, fmt.Errorf("%s: invalid token TTLs", operation)
	}

	if cfg.OpenAI.Timeout <= 0 || cfg.Server.WriteTimeout <= 0 {
		return Config{}, fmt.Errorf("%s: timeouts must be positive", operation)
	}

	if queueErr := cfg.Reports.validate(); queueErr != nil {
		return Config{}, fmt.Errorf("%s: %w", operation, queueErr)
	}

	if cfg.Environment == EnvProduction && !cfg.Auth.CookieSecure {
		return Config{}, fmt.Errorf("%s: COOKIE_SECURE is required in production", operation)
	}

	if strings.TrimSpace(cfg.Auth.AllowedOrigins) == "" {
		return Config{}, fmt.Errorf("%s: ALLOWED_ORIGINS is required", operation)
	}

	for _, origin := range strings.Split(cfg.Auth.AllowedOrigins, ",") {
		u, err := url.Parse(strings.TrimSpace(origin))
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
			return Config{}, fmt.Errorf("%s: invalid allowed origin", operation)
		}

		if cfg.Environment == EnvProduction && u.Scheme != "https" {
			return Config{}, fmt.Errorf("%s: origins must use HTTPS in production", operation)
		}
	}

	cfg.App.Version = version

	return cfg, nil
}

func (reports Reports) validate() error {
	if reports.Role != "api" && reports.Role != "worker" && reports.Role != "all" {
		return fmt.Errorf("%s: PROCESS_ROLE must be api, worker or all", "Reports.validate")
	}

	if reports.Concurrency < 1 || reports.Concurrency > 16 || reports.Stream == "" || reports.Group == "" {
		return fmt.Errorf("%s: invalid report queue configuration", "Reports.validate")
	}

	redisURL, parseErr := url.Parse(reports.RedisURL)
	if parseErr != nil || redisURL.Host == "" || (redisURL.Scheme != "redis" && redisURL.Scheme != "rediss") {
		return fmt.Errorf("%s: invalid REDIS_URL", "Reports.validate")
	}

	return nil
}
