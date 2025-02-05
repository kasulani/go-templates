package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	Logger struct {
		*zap.Logger
	}

	// ContextKey represents a context key as a string.
	ContextKey string
)

const (
	CorrelationID = ContextKey("correlationID")
)

func GetCorrelationIDFromCtx(ctx context.Context) string {
	val := ctx.Value(CorrelationID)
	if val == nil {
		return ""
	}

	id, ok := val.(string)
	if !ok {
		return ""
	}

	return id
}

// NewLogger returns an instance of Logger.
func NewLogger() *Logger {
	logLevel := os.Getenv("LOG_LEVEL")

	cfg := zap.NewProductionConfig()

	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "message"
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true

	switch logLevel {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		cfg.DisableCaller = false
		cfg.DisableStacktrace = false
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	l, err := cfg.Build()
	if err != nil {
		log.Fatalf("failed to load newLogger: %q", err)
	}

	return &Logger{l}
}

// CorrelationIDField returns a zap.Field with the correlation_id key.
func CorrelationIDField(id string) zap.Field {
	if id == "" {
		return zap.Skip()
	}

	return zap.String("correlation_id", id)
}

// SanitizeSecrets replaces sensitive information in the input string with "redacted". Add more patterns as needed.
func SanitizeSecrets(input string) string {
	secretPatterns := map[string]string{
		"password=":      `password=[^&]*`,
		"client_id=":     `client_id=[^&]*`,
		"client_secret=": `client_secret=[^&]*`,
		"username=":      `username=[^&]*`,
		"access_token:":  `"access_token":"[^"]*"`,
		"refresh_token:": `"refresh_token":"[^"]*"`,
		"username:":      `"username":"[^"]*"`,
	}

	sanitized := input
	for field, pattern := range secretPatterns {
		re := regexp.MustCompile(pattern)
		sanitized = re.ReplaceAllString(sanitized, fmt.Sprintf("%s%s", field, `"redacted"`))
	}

	return sanitized
}
