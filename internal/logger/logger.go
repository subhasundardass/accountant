package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gogpu/gogpu"
)

// level controls the minimum log level output by this package's logger.
// Using slog.LevelVar lets you change the level at runtime (e.g. via a
// debug flag, admin endpoint, or signal handler) without restarting the app.
var level = new(slog.LevelVar) // defaults to slog.LevelInfo

var log = slog.New(
	slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: level,
		},
	),
)

func init() {
	// Optional: bump to debug via env var, e.g. APP_DEBUG=1 go run .
	if os.Getenv("APP_DEBUG") == "1" {
		level.Set(slog.LevelDebug)
	}

	// Route gogpu's internal Debug/Info/Warn logs (backend selection,
	// adapter info, resource cleanup, etc.) through this same logger,
	// so app logs and framework logs share one format/output.
	gogpu.SetLogger(log)
}

// SetLevel changes the minimum log level at runtime.
func SetLevel(l slog.Level) {
	level.Set(l)
}

// With returns a child logger with the given key-value pairs attached to
// every subsequent log line. Use this to tag logs per module without
// repeating attributes on every call:
//
//	var log = logger.With("module", "ledger")
//	log.Info("entry saved", "id", e.ID)
func With(args ...any) *slog.Logger {
	return log.With(args...)
}

// --- structured, level-based logging ---

func Debug(msg string, args ...any) {
	log.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	log.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	log.Error(msg, args...)
}

// --- context-aware variants (for tracing/request-scoped logging) ---

func DebugContext(ctx context.Context, msg string, args ...any) {
	log.DebugContext(ctx, msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	log.InfoContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	log.WarnContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	log.ErrorContext(ctx, msg, args...)
}

// --- printf-style convenience variants ---
//
// NOTE: these collapse everything into a single "msg" string, which loses
// slog's structured key-value attributes. Fine for quick console debug
// traces; prefer the structured functions above for anything you might
// want to filter/query later (e.g. once using a JSON handler).

func Debugf(format string, args ...any) {
	log.Debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	log.Info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	log.Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	log.Error(fmt.Sprintf(format, args...))
}
