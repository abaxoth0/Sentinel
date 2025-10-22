package config

import (
	"io"
	"os"
	"sentinel/packages/common/logger"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

var log = logger.NewSource("CONFIG", logger.Default)

// Wrapper for time.ParseDuration. Panics on error.
func parseDuration(raw string) time.Duration {
	v, e := time.ParseDuration(raw)

	if e != nil {
		panic(e)
	}

	return v
}

type dbConfig struct {
	RawDefaultQueryTimeout string `yaml:"db-default-queuery-timeout" validate:"required"`
	MaxSearchPageSize      int    `yaml:"db-max-search-page-size" validate:"gt=0"`
	SkipPostConnection     bool   `yaml:"db-skip-post-connection" validate:"exists"`
}

func (c *dbConfig) DefaultQueryTimeout() time.Duration {
	return parseDuration(c.RawDefaultQueryTimeout)
}

type httpServerConfig struct {
	Domain         string   `yaml:"domain" validate:"required"`
	Secured        bool     `yaml:"http-secured" validate:"exists"`
	Port           string   `yaml:"http-port" validate:"required"`
	AllowedOrigins []string `yaml:"http-allowed-origins" validate:"required,min=1"`
}

type authConfing struct {
	RawAccessTokenTTL  string   `yaml:"access-token-ttl" validate:"required"`
	RawRefreshTokenTTL string   `yaml:"refresh-token-ttl" validate:"required"`
	TokenAudience      []string `yaml:"token-audience" validate:"required,min=1"`
	SelfAudience       string   `yaml:"self-audience" validate:"required"`
}

func (c *authConfing) AccessTokenTTL() time.Duration {
	return parseDuration(c.RawAccessTokenTTL)
}

func (c *authConfing) RefreshTokenTTL() time.Duration {
	return parseDuration(c.RawRefreshTokenTTL)
}

type cacheConfig struct {
	RawPoolTimeout      string `yaml:"cache-pool-timeout" validate:"required"`
	RawOperationTimeout string `yaml:"cache-operation-timeout" validate:"required"`
	RawTTL              string `yaml:"cache-ttl" validate:"required"`
}

func (c *cacheConfig) PoolTimeout() time.Duration {
	return parseDuration(c.RawPoolTimeout)
}

func (c *cacheConfig) OperationTimeout() time.Duration {
	return parseDuration(c.RawOperationTimeout)
}

func (c *cacheConfig) TTL() time.Duration {
	return parseDuration(c.RawTTL)
}

type debugConfig struct {
	Enabled           bool `yaml:"debug-mode" validate:"exists"`
	SafeDatabaseScans bool `yaml:"debug-safe-db-scans" validate:"exists"`
	LogDbQueries      bool `yaml:"debug-log-db-queries" validate:"exists"`
	// If not empty then this IP will be used to get all users location.
	// Designed for debugging, cuz location provider will return error if IP is local, e.g. 127.0.0.1
	LocationIP string `yaml:"debug-location-ip"`
}

type appConfig struct {
	ShowLogs                 bool   `yaml:"show-logs" validate:"exists"`
	TraceLogsEnabled         bool   `yaml:"trace-logs" validate:"exists"`
	ServiceID                string `yaml:"service-id" validate:"required"`
	RawActivationTokenTTL    string `yaml:"user-activation-token-ttl" validate:"required"`
	RawPasswordResetTokenTTL string `yaml:"password-reset-token-ttl" validate:"required"`
	PasswordResetRedirectURL string `yaml:"password-reset-redirect-url" validate:"required"`
}

type emailConfig struct {
	SmtpHost string `yaml:"smtp-host" validate:"required"`
	SmtpPort int    `yaml:"smtp-port" validate:"required"`
}

func (c *appConfig) ActivationTokenTTL() time.Duration {
	return parseDuration(c.RawActivationTokenTTL)
}

func (c *appConfig) PasswordResetTokenTTL() time.Duration {
	return parseDuration(c.RawPasswordResetTokenTTL)
}

type sentry struct {
	TraceSampleRate float64 `yaml:"sentry-trace-sample-rate" validate:"required,min=0.0,max=1.0"`
}

type configs struct {
	dbConfig         `yaml:",inline"`
	httpServerConfig `yaml:",inline"`
	authConfing      `yaml:",inline"`
	cacheConfig      `yaml:",inline"`
	debugConfig      `yaml:",inline"`
	appConfig        `yaml:",inline"`
	emailConfig      `yaml:",inline"`
	sentry           `yaml:",inline"`
}

var (
	DB     *dbConfig
	HTTP   *httpServerConfig
	Auth   *authConfing
	Cache  *cacheConfig
	Debug  *debugConfig
	App    *appConfig
	Email  *emailConfig
	Sentry *sentry
)

var isInit bool = false

func loadConfig(path string, dest *configs) {
	log.Info("Reading config file...", nil)

	file, err := os.Open(path)

	if err != nil {
		log.Fatal("Failed to open config file", err.Error(), nil)
	}

	rawConfig, err := io.ReadAll(file)

	if err != nil {
		log.Fatal("Failed to read config file", err.Error(), nil)
	}

	log.Info("Reading config file: OK", nil)

	log.Info("Parsing config file...", nil)

	if err := yaml.Unmarshal(rawConfig, dest); err != nil {
		log.Fatal("Failed to parse config file", err.Error(), nil)
	}

	log.Info("Parsing config file: OK", nil)

	log.Info("Validating config...", nil)

	validate := validator.New()

	validate.RegisterValidation("exists", func(fl validator.FieldLevel) bool {
		return true // Always pass (just ensure that the field exists)
	})

	if err := validate.Struct(dest); err != nil {
		log.Fatal("Failed to validate config", err.Error(), nil)
		os.Exit(1)
	}

	if !slices.Contains(dest.authConfing.TokenAudience, dest.authConfing.SelfAudience) {
		log.Fatal("Failed to validate config", "Value of 'self-audience' must exists in 'token-audience'", nil)
		os.Exit(1)
	}

	log.Info("Validating config: OK", nil)
}

func Init() {
	if isInit {
		log.Fatal("Failed to initialize config", "Config already initialized", nil)
	}

	log.Info("Initializing...", nil)

	configs := new(configs)

	loadConfig("sentinel.config.yaml", configs)
	loadSecrets()

	jwt.RegisterSigningMethod(jwt.SigningMethodEdDSA.Alg(), func() jwt.SigningMethod { return jwt.SigningMethodEdDSA })

	DB = &configs.dbConfig
	HTTP = &configs.httpServerConfig
	Auth = &configs.authConfing
	Cache = &configs.cacheConfig
	Debug = &configs.debugConfig
	App = &configs.appConfig
	Email = &configs.emailConfig
	Sentry = &configs.sentry

	log.Info("Initializing: OK", nil)

	isInit = true
}
