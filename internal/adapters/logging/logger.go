package logging

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

func toSlogLevel(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type LogCallback func(message string)

type callbackHandler struct {
	level    *slog.LevelVar
	callback LogCallback
	mu       sync.RWMutex
}

func newCallbackHandler(level Level) *callbackHandler {
	levelVar := &slog.LevelVar{}
	levelVar.Set(toSlogLevel(level))
	return &callbackHandler{
		level: levelVar,
	}
}

func (h *callbackHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *callbackHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.RLock()
	cb := h.callback
	h.mu.RUnlock()

	levelName := strings.ToUpper(r.Level.String())
	message := fmt.Sprintf("[%s] %s", levelName, r.Message)

	if cb != nil {
		cb(message)
	} else {
		fmt.Println(message)
	}

	return nil
}

func (h *callbackHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *callbackHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *callbackHandler) setCallback(cb LogCallback) {
	h.mu.Lock()
	h.callback = cb
	h.mu.Unlock()
}

func (h *callbackHandler) setLevel(level Level) {
	h.level.Set(toSlogLevel(level))
}

type Logger struct {
	slogger *slog.Logger
	handler *callbackHandler
}

func New(level Level) *Logger {
	handler := newCallbackHandler(level)
	return &Logger{
		slogger: slog.New(handler),
		handler: handler,
	}
}

func NewFromString(levelStr string) *Logger {
	return New(ParseLevel(levelStr))
}

func (l *Logger) SetLevel(level Level) {
	l.handler.setLevel(level)
}

func (l *Logger) SetCallback(cb LogCallback) {
	l.handler.setCallback(cb)
}

func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.slogger.DebugContext(ctx, fmt.Sprintf(format, args...))
}

// Infof logs an info message.
func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	l.slogger.InfoContext(ctx, fmt.Sprintf(format, args...))
}

func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.slogger.WarnContext(ctx, fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.slogger.ErrorContext(ctx, fmt.Sprintf(format, args...))
}
