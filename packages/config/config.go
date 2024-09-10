package config

import (
	"crypto/ed25519"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

type databaseConfig struct {
	Name                      string
	UserCollectionName        string
	DeletedUserCollectionName string
	Username                  string
	Password                  string
	URI                       string
	QueryDefaultTimeout       time.Duration
}

type httpServerConfig struct {
	Port string
}

type jwtConfing struct {
	AccessTokenPrivateKey  *ed25519.PrivateKey
	AccessTokenPublicKey   *ed25519.PublicKey
	RefreshTokenPrivateKey *ed25519.PrivateKey
	RefreshTokenPublicKey  *ed25519.PublicKey
	AccessTokenTTL         time.Duration
	RefreshTokenTTL        time.Duration
}

type cacheConfig struct {
	URI           string
	Password      string
	DB            int
	SocketTimeout time.Duration
	TTL           time.Duration
}

type debugConfig struct {
	Enabled bool
}

func getEnv(key string) string {
	env, _ := os.LookupEnv(key)

	log.Println("[ ENV ] Loaded: " + key)

	return env
}

var DB, HTTP, JWT, Cache, Debug = (func() (*databaseConfig, *httpServerConfig, *jwtConfing, *cacheConfig, *debugConfig) {
	log.Println("[ CONFIG ] Initializing...")

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	requiredVariables := [19]string{
		"VERSION",
		"SERVER_PORT",
		"DEBUG_ENABLED",
		"DB_NAME",
		"DB_USER_COLLECTION_NAME",
		"DB_DELETED_USER_COLLECTION_NAME",
		"DB_USER_NAME",
		"DB_PASSWORD",
		"DB_URI",
		"DB_DEFAULT_TIMEOUT",
		"ACCESS_TOKEN_SECRET",
		"REFRESH_TOKEN_SECRET",
		"ACCESS_TOKEN_TTL",
		"REFRESH_TOKEN_TTL",
		"CACHE_URI",
		"CACHE_PASSWORD",
		"CACHE_DB",
		"CACHE_SOCKET_READ_TIMEOUT",
		"CACHE_TTL",
	}

	// Check is all required env variables exists
	for _, variable := range requiredVariables {
		if _, exists := os.LookupEnv(variable); !exists {
			log.Fatalln("[ CRITICAL ERROR ] Missing required env variable: " + variable)
		}
	}

	queryTimeoutMultiplier, _ := strconv.ParseInt(getEnv("DB_DEFAULT_TIMEOUT"), 10, 64)

	DbConfig := databaseConfig{
		Name:                      getEnv("DB_NAME"),
		UserCollectionName:        getEnv("DB_USER_COLLECTION_NAME"),
		DeletedUserCollectionName: getEnv("DB_DELETED_USER_COLLECTION_NAME"),
		Username:                  getEnv("DB_USER_NAME"),
		Password:                  getEnv("DB_PASSWORD"),
		URI:                       getEnv("DB_URI"),
		QueryDefaultTimeout:       time.Second * time.Duration(queryTimeoutMultiplier),
	}

	HttpConfig := httpServerConfig{
		Port: getEnv("SERVER_PORT"),
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

	JWTConfig := jwtConfing{
		AccessTokenPrivateKey:  &AccessTokenPrivateKey,
		RefreshTokenPrivateKey: &RefreshTokenPrivateKey,
		AccessTokenPublicKey:   &AccessTokenPublicKey,
		RefreshTokenPublicKey:  &RefreshTokenPublicKey,
		AccessTokenTTL:         time.Minute * time.Duration(AccessTokenTtlMultiplier),
		RefreshTokenTTL:        time.Hour * 24 * time.Duration(RefreshTokenTtlMultiplier),
	}

	CacheDB, _ := strconv.ParseInt(getEnv("CACHE_DB"), 10, 64)
	CacheSocketTimeoutMultiplier, _ := strconv.ParseInt(getEnv("CACHE_SOCKET_READ_TIMEOUT"), 10, 64)
	CacheTTLMultiplier, _ := strconv.ParseInt(getEnv("CACHE_TTL"), 10, 64)

	CacheConfig := cacheConfig{
		URI:           getEnv("CACHE_URI"),
		Password:      getEnv("CACHE_PASSWORD"),
		DB:            int(CacheDB),
		SocketTimeout: time.Second * time.Duration(CacheSocketTimeoutMultiplier),
		TTL:           time.Minute * time.Duration(CacheTTLMultiplier),
	}

	DebugConfig := debugConfig{
		Enabled: getEnv("DEBUG_ENABLED") == "true",
	}

	log.Println("[ CONFIG ] Initializing: OK")

	return &DbConfig, &HttpConfig, &JWTConfig, &CacheConfig, &DebugConfig
})()
