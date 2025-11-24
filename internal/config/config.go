package config

import "os"

type Config struct {
	HTTPPort      string
	SecureCookies bool
	DatabaseUrl   string
	Domain	      string
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	SMTPFrom      string
	S3Endpoint    string
	S3Bucket      string
	S3AccessKey   string
	S3SecretKey   string
	S3UseSSL      bool
}

func New() *Config {
	return &Config{
		HTTPPort:      getEnv("HTTP_PORT", "80"),
		SecureCookies: getEnv("SECURE_COOKIES", "true") == "true",
		DatabaseUrl:	 getEnv("DATABASE_URL", "file:/data/sqlite.db?_journal_mode=WAL&_foreign_keys=on&_recursive_triggers=off&_busy_timeout=5000"),
		Domain:        getEnv("DOMAIN", "http://localhost:5173"),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", ""),
		SMTPUser:      getEnv("SMTP_USERNAME", ""),
		SMTPPass:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:      getEnv("SMTP_FROM", ""),
		S3Endpoint:    getEnv("S3_ENDPOINT", ""),
		S3Bucket:      getEnv("S3_BUCKET", ""),
		S3AccessKey:   getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:   getEnv("S3_SECRET_KEY", ""),
		S3UseSSL:      getEnv("S3_USE_SSL", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
