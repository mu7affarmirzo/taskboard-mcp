package config

import "testing"

func TestLoad_FromEnvVars(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "test-token-123")
	t.Setenv("TRELLO_API_KEY", "trello-key-456")
	t.Setenv("CLAUDE_API_KEY", "claude-key-789")
	t.Setenv("CLAUDE_MODEL", "claude-sonnet-4-5-20250929")
	t.Setenv("DATABASE_PATH", "/tmp/test.db")
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("TELEGRAM_MODE", "polling")

	cfg := Load()

	if cfg.TelegramToken != "test-token-123" {
		t.Errorf("expected TelegramToken 'test-token-123', got %q", cfg.TelegramToken)
	}
	if cfg.TrelloAPIKey != "trello-key-456" {
		t.Errorf("expected TrelloAPIKey 'trello-key-456', got %q", cfg.TrelloAPIKey)
	}
	if cfg.ClaudeAPIKey != "claude-key-789" {
		t.Errorf("expected ClaudeAPIKey 'claude-key-789', got %q", cfg.ClaudeAPIKey)
	}
	if cfg.ClaudeModel != "claude-sonnet-4-5-20250929" {
		t.Errorf("expected ClaudeModel 'claude-sonnet-4-5-20250929', got %q", cfg.ClaudeModel)
	}
	if cfg.DatabasePath != "/tmp/test.db" {
		t.Errorf("expected DatabasePath '/tmp/test.db', got %q", cfg.DatabasePath)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected Port '9090', got %q", cfg.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got %q", cfg.LogLevel)
	}
	if cfg.TelegramMode != "polling" {
		t.Errorf("expected TelegramMode 'polling', got %q", cfg.TelegramMode)
	}
}

func TestLoad_DefaultsWhenEmpty(t *testing.T) {
	// Clear all env vars to test defaults
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TRELLO_API_KEY", "")
	t.Setenv("CLAUDE_API_KEY", "")

	cfg := Load()

	if cfg.TelegramToken != "" {
		t.Errorf("expected empty TelegramToken, got %q", cfg.TelegramToken)
	}
	if cfg.TrelloAPIKey != "" {
		t.Errorf("expected empty TrelloAPIKey, got %q", cfg.TrelloAPIKey)
	}
}
