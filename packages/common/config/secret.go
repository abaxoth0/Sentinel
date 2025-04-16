package config

import (
	"crypto/ed25519"
	"log"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type secrets struct {
    DatabaseURI               string             `validate:"required"`
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

    log.Println("[ ENV ] Loaded: " + key)

    return env
}

func loadSecrets() {
	log.Println("[ CONFIG ] Loading environment vairables...")

    if err := godotenv.Load(); err != nil {
        log.Printf("[ CONFIG ] Failed to load environment vairables: %v\n", err)
        os.Exit(1)
    }

    requiredEnvVars := []string{
        "DB_URI",
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
            log.Fatalln("[ CRITICAL ERROR ] Missing required env variable: " + variable)
        }
    }

    cacheDB, err := strconv.ParseInt(getEnv("CACHE_DB"), 10, 64)

    if err != nil {
        log.Printf("[ CONFIG ] Failed to parse CACHE_DB env variable: %v\n", err)
        os.Exit(1)
    }

    Secret.CacheURI = getEnv("CACHE_URI")
    Secret.CachePassword = getEnv("CACHE_PASSWORD")
    Secret.CacheDB = int(cacheDB)
    Secret.MailerEmailPassword = getEnv("MAILER_EMAIL_PASSWORD")
    Secret.MailerEmail = getEnv("MAILER_EMAIL")

    Secret.DatabaseURI = getEnv("DB_URI")

    // All must be 32 bytes long
    AccessTokenSecret := []byte(getEnv("ACCESS_TOKEN_SECRET"))
    RefreshTokenSecret := []byte(getEnv("REFRESH_TOKEN_SECRET"))
    ActivationTokenSecret := []byte(getEnv("ACTIVATION_TOKEN_SECRET"))

    if len(AccessTokenSecret) != 32 {
        log.Fatalln("[ CONFIG ] Invalid length of access token secret (must be 32 bytes long)")
    }
    if len(RefreshTokenSecret) != 32 {
        log.Fatalln("[ CONFIG ] Invalid length of refresh token secret (must be 32 bytes long)")
    }
    if len(ActivationTokenSecret) != 32 {
        log.Fatalln("[ CONFIG ] Invalid length of activation token secret (must be 32 bytes long)")
    }

    Secret.AccessTokenPrivateKey = ed25519.NewKeyFromSeed(AccessTokenSecret)
    Secret.RefreshTokenPrivateKey = ed25519.NewKeyFromSeed(RefreshTokenSecret)
    Secret.ActivationTokenPrivateKey = ed25519.NewKeyFromSeed(ActivationTokenSecret)

    // `priv.Public()` actually returns `ed25519.PublicKey` type, not `crypto.PublicKey`.
    // Tested via `reflect.TypeOf()`
    Secret.AccessTokenPublicKey = Secret.AccessTokenPrivateKey.Public().(ed25519.PublicKey)
    Secret.RefreshTokenPublicKey = Secret.RefreshTokenPrivateKey.Public().(ed25519.PublicKey)
    Secret.ActivationTokenPublicKey = Secret.ActivationTokenPrivateKey.Public().(ed25519.PublicKey)

    log.Println("[ CONFIG ] Loading environment vairables: OK")

    log.Println("[ CONFIG ] Validating secrets...")


    validate := validator.New()

    validate.RegisterValidation("exists", func(fl validator.FieldLevel) bool {
        return true // Always pass (just ensure that the field exists)
    })

    if err := validate.Struct(Secret); err != nil {
        log.Printf("[ CONFIG ] Secrets validation failed: %v\n", err)
        os.Exit(1)
    }

    log.Println("[ CONFIG ] Validating secrets: OK")
}

