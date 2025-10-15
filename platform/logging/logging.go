package logging

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	LevelInfo  Level = 1
	LevelWarn  Level = 2
	LevelError Level = 3
)

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger interface {
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, msg string)
	SetWriter(w io.Writer)
	SetLevel(level Level)
}

type logger struct {
	mu     sync.Mutex
	writer io.Writer
	level  Level
}

func NewLogger() Logger {
	return &logger{
		writer: os.Stdout,
		level:  LevelInfo,
	}
}

func (l *logger) Info(ctx context.Context, msg string) {
	l.log(ctx, LevelInfo, msg)
}

func (l *logger) Warn(ctx context.Context, msg string) {
	l.log(ctx, LevelWarn, msg)
}

func (l *logger) Error(ctx context.Context, msg string) {
	l.log(ctx, LevelError, msg)
}

func (l *logger) SetWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if w == nil {
		l.writer = os.Stdout
		return
	}

	l.writer = w
}

func (l *logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < LevelInfo {
		level = LevelInfo
	}
	if level > LevelError {
		level = LevelError
	}

	l.level = level
}

func (l *logger) log(_ context.Context, level Level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	fmt.Fprintf(l.writer, "%s [%s] %s\n", timestamp, level.String(), msg)
}

var (
	defaultLogger = NewLogger()
	ctxKey        = &struct{}{}
)

func WithLogger(ctx context.Context, logger Logger) context.Context {
	if logger == nil {
		return ctx
	}

	return context.WithValue(ctx, ctxKey, logger)
}

func FromContext(ctx context.Context) Logger {
	if ctx == nil {
		return defaultLogger
	}

	if logger, ok := ctx.Value(ctxKey).(Logger); ok && logger != nil {
		return logger
	}

	return defaultLogger
}
