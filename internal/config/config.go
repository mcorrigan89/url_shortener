package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port      int
	Env       string
	ClientURL string
	DB        struct {
		DSN     string
		Logging bool
	}
	Cors struct {
		TrustedOrigins []string
	}
	OAuth struct {
		Google struct {
			ClientID     string
			ClientSecret string
		}
	}
}

func LoadConfig(cfg *Config) {

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Load ENV
	env := os.Getenv("ENV")
	if env == "" {
		cfg.Env = "local"
	} else {
		cfg.Env = env
	}

	// Load PORT
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("PORT not available in .env")
	}
	cfg.Port = port

	// Load CLIENT_URL
	client_url := os.Getenv("CLIENT_URL")
	if client_url == "" {
		log.Fatalf("CLIENT_URL not available in .env")
	}
	cfg.ClientURL = client_url

	// Load DATABASE_URL
	postgres_url := os.Getenv("POSTGRES_URL")
	if postgres_url == "" {
		log.Fatalf("POSTGRES_URL not available in .env")
	}
	cfg.DB.DSN = strings.Replace(postgres_url, "?schema=public", "", -1)

	cfg.Cors.TrustedOrigins = []string{"http://localhost:3000", "https://url.corrigan.io"}

	// Load OAuth
	google_client_id := os.Getenv("GOOGLE_CLIENT_ID")
	if google_client_id == "" {
		log.Fatalf("GOOGLE_CLIENT_ID not available in .env")
	}

	google_client_secret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if google_client_secret == "" {
		log.Fatalf("GOOGLE_CLIENT_SECRET not available in .env")
	}

	cfg.OAuth.Google.ClientID = google_client_id
	cfg.OAuth.Google.ClientSecret = google_client_secret
}
