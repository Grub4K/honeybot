package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rs/zerolog"
)

var Level = &slog.LevelVar{}

func init() {
	zerolog.MessageFieldName = slog.MessageKey
	zerolog.CallerFieldName = slog.SourceKey
	slog.SetDefault(slog.New(slog.NewJSONHandler(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05 Z07:00",
	}, &slog.HandlerOptions{
		Level:     Level,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if groups == nil && a.Key == slog.SourceKey {
				source, ok := a.Value.Any().(*slog.Source)
				if !ok {
					return a
				}
				value := fmt.Sprintf("%s:%d", source.File, source.Line)
				return slog.String(slog.SourceKey, value)
			}
			return a
		},
	})))
}
