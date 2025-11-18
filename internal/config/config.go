package config

import "os"

type Config struct {
	DBPath        string
	HTTPPort      string
	SecureCookies bool
	AppDSN        string
	GooseDSN      string
}

func New() *Config {
	dbPath := getEnv("DB_PATH", "dev.db")

	return &Config{
		DBPath:        dbPath,
		HTTPPort:      getEnv("HTTP_PORT", "8080"),
		SecureCookies: getEnv("SECURE_COOKIES", "true") == "true",
		GooseDSN:      "file:" + dbPath,
		AppDSN:        "file:" + dbPath + "?_journal_mode=WAL&_foreign_keys=on&_recursive_triggers=off",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
