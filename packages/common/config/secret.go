package config

import (
	"crypto/ed25519"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type secrets struct {
    DatabaseHost              string             `validate:"required"`
    DatabasePort              string             `validate:"required"`
    DatabaseName              string             `validate:"required"`
    DatabaseUser              string             `validate:"required"`
    DatabasePassword          string             `validate:"required"`
    AccessTokenPrivateKey     ed25519.PrivateKey `validate:"required"`
    AccessTokenPublicKey      ed25519.PublicKey  `validate:"required"`
    RefreshTokenPrivateKey    ed25519.PrivateKey `validate:"required"`
    RefreshTokenPublicKey     ed25519.PublicKey  `validate:"required"`
    ActivationTokenPrivateKey ed25519.PrivateKey `validate:"required"`
    ActivationTokenPublicKey  ed25519.PublicKey `validate:"required"`
    CacheURI                  string             `validate:"required"`
    CachePassword             string             `validate:"required"`
    CacheDB                   int                `validate:"exists"`
    MailerEmailPassword       string             `validate:"required"`
    MailerEmail               string             `validate:"required"`
}

var Secret secrets

func getEnv(key string) string {
    env, _ := os.LookupEnv(key)

    configLogger.Info("Loaded: " + key, nil)

    return env
}

func loadSecrets() {
	configLogger.Info("Loading environment vairables...", nil)

    if err := godotenv.Load(); err != nil {
        configLogger.Fatal("Failed to load environment vairables", err.Error(), nil)
    }

    requiredEnvVars := []string{
        "DB_HOST",
        "DB_PORT",
        "DB_NAME",
        "DB_USER",
        "DB_PASSWORD",
        "ACCESS_TOKEN_SECRET",
        "REFRESH_TOKEN_SECRET",
        "ACTIVATION_TOKEN_SECRET",
        "CACHE_URI",
        "CACHE_PASSWORD",
        "CACHE_DB",
        "MAILER_EMAIL_PASSWORD",
        "MAILER_EMAIL",
    }

    // Check is all required env variables exists
    for _, variable := range requiredEnvVars {
        if _, exists := os.LookupEnv(variable); !exists {
            configLogger.Fatal(
                "Failed to load environment variables",
                "Missing required env variable" + variable,
            	nil,
			)
        }
    }

    cacheDB, err := strconv.ParseInt(getEnv("CACHE_DB"), 10, 64)

    if err != nil {
        configLogger.Fatal("Failed to parse CACHE_DB env variable", err.Error(), nil)
    }

    Secret.CacheURI = getEnv("CACHE_URI")
    Secret.CachePassword = getEnv("CACHE_PASSWORD")
    Secret.CacheDB = int(cacheDB)
    Secret.MailerEmailPassword = getEnv("MAILER_EMAIL_PASSWORD")
    Secret.MailerEmail = getEnv("MAILER_EMAIL")

    Secret.DatabaseHost = getEnv("DB_HOST")
    Secret.DatabasePort = getEnv("DB_PORT")
    Secret.DatabaseName = getEnv("DB_NAME")
    Secret.DatabaseUser = getEnv("DB_USER")
    Secret.DatabasePassword = getEnv("DB_PASSWORD")

    // All must be 32 bytes long
    AccessTokenSecret := []byte(getEnv("ACCESS_TOKEN_SECRET"))
    RefreshTokenSecret := []byte(getEnv("REFRESH_TOKEN_SECRET"))
    ActivationTokenSecret := []byte(getEnv("ACTIVATION_TOKEN_SECRET"))

    if len(AccessTokenSecret) != 32 {
        configLogger.Fatal(
            "Invalid environment variable value",
            "Invalid length of access token secret (must be 32 bytes long)",
         	nil,
		)
    }
    if len(RefreshTokenSecret) != 32 {
        configLogger.Fatal(
            "Invalid environment variable value",
            "Invalid length of refresh token secret (must be 32 bytes long)",
			nil,
        )
    }
    if len(ActivationTokenSecret) != 32 {
        configLogger.Fatal(
            "Invalid environment variable value",
            "Invalid length of activation token secret (must be 32 bytes long)",
			nil,
        )
    }

    Secret.AccessTokenPrivateKey = ed25519.NewKeyFromSeed(AccessTokenSecret)
    Secret.RefreshTokenPrivateKey = ed25519.NewKeyFromSeed(RefreshTokenSecret)
    Secret.ActivationTokenPrivateKey = ed25519.NewKeyFromSeed(ActivationTokenSecret)

    // `priv.Public()` actually returns `ed25519.PublicKey` type, not `crypto.PublicKey`.
    // Tested via `reflect.TypeOf()`
    Secret.AccessTokenPublicKey = Secret.AccessTokenPrivateKey.Public().(ed25519.PublicKey)
    Secret.RefreshTokenPublicKey = Secret.RefreshTokenPrivateKey.Public().(ed25519.PublicKey)
    Secret.ActivationTokenPublicKey = Secret.ActivationTokenPrivateKey.Public().(ed25519.PublicKey)

    configLogger.Info("Loading environment vairables: OK", nil)

    configLogger.Info("Validating secrets...", nil)

    validate := validator.New()

    validate.RegisterValidation("exists", func(fl validator.FieldLevel) bool {
        return true // Always pass (just ensure that the field exists)
    })

    if err := validate.Struct(Secret); err != nil {
        configLogger.Fatal("Secrets validation failed", err.Error(), nil)
    }

    configLogger.Info("Validating secrets: OK", nil)
}

