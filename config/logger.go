package config

import (
	"io"
	"log/slog"
	"time"
)

// replaceAttrFormat is a function to replace the time attribute
func replaceAttrFormat(group []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		t := time.Now().Format("2006-01-02T15:04:05")
		return  slog.Attr{Key: a.Key, Value: slog.StringValue(t)}
	}
	return  a
}

func NewLogger(out io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(
		out,
		&slog.HandlerOptions{
			AddSource: true,
			Level: level,
			ReplaceAttr: replaceAttrFormat,
		},
	))
}