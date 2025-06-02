package config

import (
	"io"
	"os"
	"sentinel/packages/common/logger"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/yaml.v3"
)

var configLogger = logger.NewSource("CONFIG",logger.Default)

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
}

func (c * dbConfig) DefaultQueryTimeout() time.Duration {
    return parseDuration(c.RawDefaultQueryTimeout)
}

type httpServerConfig struct {
    Domain         string   `yaml:"domain" validate:"required"`
    Secured        bool     `yaml:"http-secured" validate:"exists"`
    Port           string   `yaml:"http-port" validate:"required"`
    AllowedOrigins []string `yaml:"http-allowed-origins" validate:"required,min=1"`
}

type authConfing struct {
    RawAccessTokenTTL  string `yaml:"access-token-ttl" validate:"required"`
    RawRefreshTokenTTL string `yaml:"refresh-token-ttl" validate:"required"`
}

func (c *authConfing) AccessTokenTTL() time.Duration {
    return parseDuration(c.RawAccessTokenTTL)
}

func (c *authConfing) RefreshTokenTTL() time.Duration {
    return parseDuration(c.RawRefreshTokenTTL)
}

type cacheConfig struct {
    RawSocketTimeout    string `yaml:"cache-socket-timeout" validate:"required"`
    RawOperationTimeout string `yaml:"cache-operation-timeout" validate:"required"`
    RawTTL              string `yaml:"cache-ttl" validate:"required"`
}

func (c *cacheConfig) SocketTimeout() time.Duration {
    return parseDuration(c.RawSocketTimeout)
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
}

type appConfig struct {
    ShowLogs              bool   `yaml:"show-logs" validate:"exists"`
    TraceLogsEnabled      bool   `yaml:"trace-logs" validate:"exists"`
    ServiceID             string `yaml:"service-id" validate:"required"`
    IsLoginEmail          bool   `yaml:"is-login-email" validate:"exists"`
    RawActivationTokenTTL string `yaml:"user-activation-token-ttl" validate:"required"`
}

type emailConfig struct {
    SmtpHost    string `yaml:"smtp-host" validate:"required"`
    SmtpPort    int    `yaml:"smtp-port" validate:"required"`
}

func (c *appConfig) ActivationTokenTTL() time.Duration {
    return parseDuration(c.RawActivationTokenTTL)
}

type configs struct {
    dbConfig         `yaml:",inline"`
    httpServerConfig `yaml:",inline"`
    authConfing      `yaml:",inline"`
    cacheConfig      `yaml:",inline"`
    debugConfig      `yaml:",inline"`
    appConfig        `yaml:",inline"`
    emailConfig      `yaml:",inline"`
}

var DB *dbConfig
var HTTP *httpServerConfig
var Auth *authConfing
var Cache *cacheConfig
var Debug *debugConfig
var App *appConfig
var Email *emailConfig

var isInit bool = false

func loadConfig(path string, dest *configs) {
	configLogger.Info("Reading config file...")

    file, err := os.Open(path)

    if err != nil {
        configLogger.Fatal("Failed to open config file", err.Error())
    }

    rawConfig, err := io.ReadAll(file)

    if err != nil {
        configLogger.Fatal("Failed to read config file", err.Error())
    }

    configLogger.Info("Reading config file: OK")

    configLogger.Info("Parsing config file...")

    if err := yaml.Unmarshal(rawConfig, dest); err != nil {
        configLogger.Fatal("Failed to parse config file", err.Error())
    }

    configLogger.Info("Parsing config file: OK")

    configLogger.Info("Validating config...")

    validate := validator.New()

    validate.RegisterValidation("exists", func(fl validator.FieldLevel) bool {
        return true // Always pass (just ensure that the field exists)
    })

    if err := validate.Struct(dest); err != nil {
        configLogger.Fatal("Failed to validate config", err.Error())
        os.Exit(1)
    }

    configLogger.Info("Validating config: OK")
}

func Init() {
    if isInit {
        configLogger.Fatal("Failed to initialize config", "Config already initialized")
    }

	configLogger.Info("Initializing...")

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

	configLogger.Info("Initializing: OK")

    isInit = true
}

