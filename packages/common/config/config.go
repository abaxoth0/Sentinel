package config

import (
	"io"
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"gopkg.in/yaml.v3"
)

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
    Secured        bool     `yaml:"http-secured" validate:"exists"`
    Port           string   `yaml:"http-port" validate:"required"`
    AllowedOrigins []string `yaml:"http-allowed-origins" validate:"required,min=1"`
}

type jwtConfing struct {
    RawAccessTokenTTL  string `yaml:"access-token-ttl" validate:"required"`
    RawRefreshTokenTTL string `yaml:"refresh-token-ttl" validate:"required"`
}

func (c *jwtConfing) AccessTokenTTL() time.Duration {
    return parseDuration(c.RawAccessTokenTTL)
}

func (c *jwtConfing) RefreshTokenTTL() time.Duration {
    return parseDuration(c.RawRefreshTokenTTL)
}

type authorizationConfig struct {
    ServiceID string `yaml:"service-id" validate:"required"`
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
    IsLoginEmail bool `yaml:"is-login-email" validate:"exists"`
    RawActivationTokenTTL string `yaml:"user-activation-token-ttl" validate:"required"`
}

func (c *appConfig) ActivationTokenTTL() time.Duration {
    return parseDuration(c.RawActivationTokenTTL)
}

type configs struct {
    dbConfig `yaml:",inline"`
    httpServerConfig `yaml:",inline"`
    jwtConfing `yaml:",inline"`
    authorizationConfig `yaml:",inline"`
    cacheConfig `yaml:",inline"`
    debugConfig `yaml:",inline"`
    appConfig `yaml:",inline"`
}

var DB *dbConfig
var HTTP *httpServerConfig
var JWT *jwtConfing
var Authorization *authorizationConfig
var Cache *cacheConfig
var Debug *debugConfig
var App *appConfig

var isInit bool = false

func loadConfig(path string, dest *configs) {
	log.Println("[ CONFIG ] Reading config file...")

    file, err := os.Open(path)

    if err != nil {
        log.Printf("[ CONFIG ] Failed to open config file: %v\n", err)
        os.Exit(1)
    }

    rawConfig, err := io.ReadAll(file)

    if err != nil {
        log.Printf("[ CONFIG ] Failed to read config file: %v\n", err)
        os.Exit(1)
    }

    log.Println("[ CONFIG ] Reading config file: OK")

    log.Println("[ CONFIG ] Parsing config file...")

    if err := yaml.Unmarshal(rawConfig, dest); err != nil {
        log.Printf("[ CONFIG ] Failed to parse config file: %v\n", err)
        os.Exit(1)
    }

    log.Println("[ CONFIG ] Parsing config file: OK")

    log.Println("[ CONFIG ] Validating config...")

    validate := validator.New()

    validate.RegisterValidation("exists", func(fl validator.FieldLevel) bool {
        return true // Always pass (just ensure that the field exists)
    })

    if err := validate.Struct(dest); err != nil {
        log.Printf("[ CONFIG ] Failed to validate config: %v\n", err)
        os.Exit(1)
    }

    log.Println("[ CONFIG ] Validating config: OK")
}

func Init() {
    if isInit {
        log.Fatalln("[ CONFIG ] Fatal error: already initialized")
    }

	log.Println("[ CONFIG ] Initializing...")

    configs := new(configs)

    loadConfig("sentinel.config.yaml", configs)
    loadSecrets()

	jwt.RegisterSigningMethod(jwt.SigningMethodEdDSA.Alg(), func() jwt.SigningMethod { return jwt.SigningMethodEdDSA })

    DB = &configs.dbConfig
    HTTP = &configs.httpServerConfig
    JWT = &configs.jwtConfing
    Authorization = &configs.authorizationConfig
    Cache = &configs.cacheConfig
    Debug = &configs.debugConfig
    App = &configs.appConfig

	log.Println("[ CONFIG ] Initializing: OK")

    isInit = true
}

