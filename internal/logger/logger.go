package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/satrunjis/user-service/internal/constants"
	"github.com/satrunjis/user-service/internal/logger/debuglogger"
)

func New(level string, writer io.Writer) *slog.Logger {
	if writer == nil {
		writer = os.Stdout
	}
	fmt.Printf("Initializing logger with level: %s\n", level)
	var handler slog.Handler
	switch level {
	case constants.Staging:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})
	case constants.Development:
		handler = debuglogger.NewDevelopmentHandler(
			writer,
			&slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		)
	case constants.Production:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		slog.Error("Unknown logging level. Using production defaults", "input_level", level)
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	return slog.New(handler)

}
