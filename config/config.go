package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

var cfg *Config

func init() {
	initConfig()
}

func initConfig() {
	// Check if the code is running on Vercel
	if os.Getenv("VERCEL") != "1" {
		// Load environment variables from .env file if not vercel
		if err := godotenv.Load(); err != nil {
			panic(fmt.Errorf("error loading .env file: %w", err))
		}
	}

	// TODO: Auto parse to struct
	cfg = &Config{
		VercelUrl:              os.Getenv("VERCEL_URL"),
		Port:                   os.Getenv("PORT"),
		TelegramBotToken:       os.Getenv("TELEGRAM_BOT_TOKEN"),
		GsheetID:               os.Getenv("GSHEET_ID"),
		GsheetProjectID:        os.Getenv("GSHEET_PROJECT_ID"),
		GsheetUserPrivateKeyID: os.Getenv("GSHEET_USER_PRIVATE_KEY_ID"),
		GsheetUserPrivateKey:   os.Getenv("GSHEET_USER_PRIVATE_KEY"),
		GsheetUserClientEmail:  os.Getenv("GSHEET_USER_CLIENT_EMAIL"),
		GsheetUserClientID:     os.Getenv("GSHEET_USER_CLIENT_ID"),
	}
}

func GetConfig() *Config {
	if cfg == nil {
		initConfig()
	}

	return cfg
}
