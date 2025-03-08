package config

import (
	"crypto/ed25519"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

type dbConfig struct {
    URI                 string
    DefaultQueryTimeout time.Duration
}

type httpServerConfig struct {
	Port          string
	AllowedOrigins []string
    Secured       bool
}

type jwtConfing struct {
	AccessTokenPrivateKey  ed25519.PrivateKey
	AccessTokenPublicKey   ed25519.PublicKey
	RefreshTokenPrivateKey ed25519.PrivateKey
	RefreshTokenPublicKey  ed25519.PublicKey
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
}

type authorizationConfig struct {
    ServiceID string
}

type cacheConfig struct {
    URI              string
    Password         string
    DB               int
    SocketTimeout    time.Duration
    TTL              time.Duration
    OperationTimeout time.Duration
}

type debugConfig struct {
	Enabled bool
}

func getEnv(key string) string {
	env, _ := os.LookupEnv(key)

	log.Println("[ ENV ] Loaded: " + key)

	return env
}

var DB dbConfig
var HTTP httpServerConfig
var JWT jwtConfing
var Authorization authorizationConfig
var Cache cacheConfig
var Debug debugConfig

var isInit bool = false

func Init() {
    if isInit {
        panic("configs was already initialized")
    }

	log.Println("[ CONFIG ] Initializing...")

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	requiredVariables := [17]string{
		"SERVER_PORT",
		"HTTP_ALLOWED_ORIGINS",
        "HTTP_SECURED",
		"DEBUG_ENABLED",
		"DB_URI",
		"DB_DEFAULT_TIMEOUT",
		"ACCESS_TOKEN_SECRET",
		"REFRESH_TOKEN_SECRET",
		"ACCESS_TOKEN_TTL",
		"REFRESH_TOKEN_TTL",
        "SERVICE_ID",
		"CACHE_URI",
		"CACHE_PASSWORD",
		"CACHE_DB",
		"CACHE_SOCKET_TIMEOUT",
		"CACHE_TTL",
        "CACHE_OPERATION_TIMEOUT",
	}

	// Check is all required env variables exists
	for _, variable := range requiredVariables {
		if _, exists := os.LookupEnv(variable); !exists {
			log.Fatalln("[ CRITICAL ERROR ] Missing required env variable: " + variable)
		}
	}

	queryTimeoutMultiplier, _ := strconv.ParseInt(getEnv("DB_DEFAULT_TIMEOUT"), 10, 64)

	DB = dbConfig{
		URI:                 getEnv("DB_URI"),
		DefaultQueryTimeout: time.Second * time.Duration(queryTimeoutMultiplier),
	}

	HTTP = httpServerConfig{
		Port:          getEnv("SERVER_PORT"),
		AllowedOrigins: strings.Split(getEnv("HTTP_ALLOWED_ORIGINS"), ","),
        Secured: getEnv("HTTP_SECURED") == "true",
	}

	jwt.RegisterSigningMethod(jwt.SigningMethodEdDSA.Alg(), func() jwt.SigningMethod { return jwt.SigningMethodEdDSA })

	AccessTokenTtlMultiplier, _ := strconv.ParseInt(getEnv("ACCESS_TOKEN_TTL"), 10, 64)
	RefreshTokenTtlMultiplier, _ := strconv.ParseInt(getEnv("REFRESH_TOKEN_TTL"), 10, 64)

	// Both must be 32 bytes long
	AccessTokenSecret := []byte(getEnv("ACCESS_TOKEN_SECRET"))
	RefreshTokenSecret := []byte(getEnv("REFRESH_TOKEN_SECRET"))

	AccessTokenPrivateKey := ed25519.NewKeyFromSeed(AccessTokenSecret)
	RefreshTokenPrivateKey := ed25519.NewKeyFromSeed(RefreshTokenSecret)

	// `priv.Public()` actually returns `ed25519.PublicKey` type, not `crypto.PublicKey`.
	// Tested via `reflect.TypeOf()`
	AccessTokenPublicKey := AccessTokenPrivateKey.Public().(ed25519.PublicKey)
	RefreshTokenPublicKey := RefreshTokenPrivateKey.Public().(ed25519.PublicKey)

	JWT = jwtConfing{
		AccessTokenPrivateKey:  AccessTokenPrivateKey,
		RefreshTokenPrivateKey: RefreshTokenPrivateKey,
		AccessTokenPublicKey:   AccessTokenPublicKey,
		RefreshTokenPublicKey:  RefreshTokenPublicKey,
		AccessTokenTTL:         time.Minute * time.Duration(AccessTokenTtlMultiplier),
		RefreshTokenTTL:        time.Hour * 24 * time.Duration(RefreshTokenTtlMultiplier),
	}

    Authorization = authorizationConfig{
        ServiceID: getEnv("SERVICE_ID"),
    }

	CacheDB, _ := strconv.ParseInt(getEnv("CACHE_DB"), 10, 64)
	CacheSocketTimeoutMultiplier, _ := strconv.ParseInt(getEnv("CACHE_SOCKET_TIMEOUT"), 10, 64)
	CacheTTLMultiplier, _ := strconv.ParseInt(getEnv("CACHE_TTL"), 10, 64)
    CacheOperationTimeoutMultiplier, _ := strconv.ParseInt(getEnv("CACHE_OPERATION_TIMEOUT"), 10, 64)

	Cache = cacheConfig{
		URI:           getEnv("CACHE_URI"),
		Password:      getEnv("CACHE_PASSWORD"),
		DB:            int(CacheDB),
		SocketTimeout: time.Second * time.Duration(CacheSocketTimeoutMultiplier),
		TTL:           time.Minute * time.Duration(CacheTTLMultiplier),
	    OperationTimeout: time.Second * time.Duration(CacheOperationTimeoutMultiplier),
    }

	Debug = debugConfig{
		Enabled: getEnv("DEBUG_ENABLED") == "true",
	}

	log.Println("[ CONFIG ] Initializing: OK")

    isInit = true
}

