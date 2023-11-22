package config

// Configuration struct to hold environment variables
type Config struct {
	VercelUrl              string `env:"VERCEL_URL"`
	Port                   string `env:"PORT"`
	TelegramBotToken       string `env:"TELEGRAM_BOT_TOKEN"`
	GsheetID               string `env:"GSHEET_ID"`
	GsheetProjectID        string `env:"GSHEET_PROJECT_ID"`
	GsheetUserPrivateKeyID string `env:"GSHEET_USER_PRIVATE_KEY_ID"`
	GsheetUserPrivateKey   string `env:"GSHEET_USER_PRIVATE_KEY"`
	GsheetUserClientEmail  string `env:"GSHEET_USER_CLIENT_EMAIL"`
	GsheetUserClientID     string `env:"GSHEET_USER_CLIENT_ID"`
}
