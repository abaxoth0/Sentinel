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
	Name                string
	UserCollectionName  string
	Username            string
	Password            string
	URI                 string
	QueryDefaultTimeout time.Duration
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

type debugConfig struct {
	// Must be false only on project deployment.
	Enabled bool
}

func initializeConfigs() (string, *databaseConfig, *httpServerConfig, *jwtConfing, *debugConfig) {
	log.Println("[ CONFIG ] Initializing...")

	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	requiredVariables := [13]string{
		"VERSION",
		"SERVER_PORT",
		"DEBUG_ENABLED",
		"DB_NAME",
		"DB_USER_COLLECTION_NAME",
		"DB_USER_NAME",
		"DB_PASSWORD",
		"DB_URI",
		"DB_DEFAULT_TIMEOUT",
		"ACCESS_TOKEN_SECRET",
		"REFRESH_TOKEN_SECRET",
		"ACCESS_TOKEN_TTL",
		"REFRESH_TOKEN_TTL",
	}

	// Check is all required env variables exists
	for _, variable := range requiredVariables {
		if _, exists := os.LookupEnv(variable); !exists {
			log.Fatalln("[ CRITICAL ERROR ] Missing required env variable: " + variable)
		}
	}

	queryTimeoutMultiplier, _ := strconv.ParseInt(getEnv("DB_DEFAULT_TIMEOUT"), 10, 64)

	DbConfig := databaseConfig{
		Name:                getEnv("DB_NAME"),
		UserCollectionName:  getEnv("DB_USER_COLLECTION_NAME"),
		Username:            getEnv("DB_USER_NAME"),
		Password:            getEnv("DB_PASSWORD"),
		URI:                 getEnv("DB_URI"),
		QueryDefaultTimeout: time.Second * time.Duration(queryTimeoutMultiplier),
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

	DebugConfig := debugConfig{
		Enabled: getEnv("DEBUG_ENABLED") == "true",
	}

	v := getEnv("VERSION")

	log.Println("[ CONFIG ] Initializing: OK")

	return v, &DbConfig, &HttpConfig, &JWTConfig, &DebugConfig
}

func getEnv(key string) string {
	env, _ := os.LookupEnv(key)

	log.Println("[ ENV ] Loaded: " + key)

	return env
}

var AppVersion, DB, HTTP, JWT, Debug = initializeConfigs()
