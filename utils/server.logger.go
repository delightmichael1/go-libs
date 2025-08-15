package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type ServerLogger struct {
	zerolog.Logger
}

func NewLogger() *ServerLogger {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i any) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatFieldName = func(i any) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i any) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatErrFieldName = func(i any) string {
		return fmt.Sprintf("%s: ", i)
	}
	zerolog := zerolog.New(output).With().Caller().Timestamp().Logger()
	return &ServerLogger{zerolog}
}

func (l *ServerLogger) LogInfo() *zerolog.Event {
	return l.Logger.Info()
}

func (l *ServerLogger) LogError() *zerolog.Event {
	return l.Logger.Error()
}

func (l *ServerLogger) LogDebug() *zerolog.Event {
	return l.Logger.Debug()
}

// func (l *ServerLogger) LogWarn() *zerolog.Event {
// 	return l.Logger.Warn()
// }

func (l *ServerLogger) LogFatal() *zerolog.Event {
	return l.Logger.Fatal()
}

func (l *ServerLogger) StackTrace() *zerolog.Event {
	return l.Error().Stack()
}
