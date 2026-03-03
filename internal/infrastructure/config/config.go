package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramToken      string
	TelegramWebhookURL string
	TelegramMode       string
	TrelloAPIKey       string
	ClaudeAPIKey       string
	ClaudeModel        string
	DatabasePath       string
	Port               string
	LogLevel           string
	MiniAppEnabled     bool
	MiniAppURL         string
	JWTSecret          string
}

func Load() *Config {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("no .env file found, using environment variables: %v", err)
	}
	return &Config{
		TelegramToken:      viper.GetString("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookURL: viper.GetString("TELEGRAM_WEBHOOK_URL"),
		TelegramMode:       viper.GetString("TELEGRAM_MODE"),
		TrelloAPIKey:       viper.GetString("TRELLO_API_KEY"),
		ClaudeAPIKey:       viper.GetString("CLAUDE_API_KEY"),
		ClaudeModel:        viper.GetString("CLAUDE_MODEL"),
		DatabasePath:       viper.GetString("DATABASE_PATH"),
		Port:               viper.GetString("PORT"),
		LogLevel:           viper.GetString("LOG_LEVEL"),
		MiniAppEnabled:     viper.GetBool("MINIAPP_ENABLED"),
		MiniAppURL:         viper.GetString("MINIAPP_URL"),
		JWTSecret:          viper.GetString("JWT_SECRET"),
	}
}
