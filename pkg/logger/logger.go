package logger

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// CustomTextFormatter provides colorful, human-readable log formatting
type CustomTextFormatter struct {
	TimestampFormat string
	ForceColors     bool
	DisableColors   bool
	FullTimestamp   bool
}

// ANSI color codes
const (
	reset      = "\033[0m"
	bold       = "\033[1m"
	dim        = "\033[2m"
	red        = "\033[31m"
	green      = "\033[32m"
	yellow     = "\033[33m"
	blue       = "\033[34m"
	magenta    = "\033[35m"
	cyan       = "\033[36m"
	white      = "\033[37m"
	bgRed      = "\033[41m"
	bgYellow   = "\033[43m"
	bgBlue     = "\033[44m"
	lightGray  = "\033[90m"
	lightCyan  = "\033[96m"
)

// Format renders a single log entry
func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b strings.Builder

	// Check if we should use colors
	useColors := f.ForceColors || (!f.DisableColors && isTerminal())

	// Timestamp
	if f.FullTimestamp {
		timestampFormat := f.TimestampFormat
		if timestampFormat == "" {
			timestampFormat = "2006-01-02 15:04:05"
		}
		if useColors {
			b.WriteString(lightGray)
		}
		b.WriteString(entry.Time.Format(timestampFormat))
		if useColors {
			b.WriteString(reset)
		}
		b.WriteString(" ")
	}

	// Log level with color and icon
	if useColors {
		b.WriteString(getLevelIcon(entry.Level))
		b.WriteString(" ")
		b.WriteString(getLevelColor(entry.Level))
		b.WriteString(bold)
	}
	b.WriteString("[")
	b.WriteString(strings.ToUpper(entry.Level.String()))
	b.WriteString("]")
	if useColors {
		b.WriteString(reset)
	}
	b.WriteString(" ")

	// Caller information (if available)
	if entry.HasCaller() {
		funcName := extractFunctionName(entry.Caller.Function)
		fileName := extractFileName(entry.Caller.File)

		if useColors {
			b.WriteString(dim)
			b.WriteString(cyan)
		}
		b.WriteString(fmt.Sprintf("[%s:%d %s] ", fileName, entry.Caller.Line, funcName))
		if useColors {
			b.WriteString(reset)
		}
	}

	// Message
	if useColors {
		b.WriteString(white)
	}
	b.WriteString(entry.Message)
	if useColors {
		b.WriteString(reset)
	}

	// Fields
	if len(entry.Data) > 0 {
		b.WriteString(" ")
		if useColors {
			b.WriteString(dim)
		}
		b.WriteString("{")

		first := true
		for k, v := range entry.Data {
			if !first {
				b.WriteString(", ")
			}
			first = false

			if useColors {
				b.WriteString(lightCyan)
			}
			b.WriteString(k)
			if useColors {
				b.WriteString(reset)
				b.WriteString(dim)
			}
			b.WriteString("=")

			// Format value with appropriate color
			if useColors {
				b.WriteString(formatValue(v))
			} else {
				b.WriteString(fmt.Sprintf("%v", v))
			}
		}

		b.WriteString("}")
		if useColors {
			b.WriteString(reset)
		}
	}

	b.WriteString("\n")
	return []byte(b.String()), nil
}

// getLevelIcon returns an emoji/icon for the log level
func getLevelIcon(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "🔍"
	case logrus.InfoLevel:
		return "ℹ️ "
	case logrus.WarnLevel:
		return "⚠️ "
	case logrus.ErrorLevel:
		return "❌"
	case logrus.FatalLevel, logrus.PanicLevel:
		return "💀"
	default:
		return "  "
	}
}

// getLevelColor returns the color code for the log level
func getLevelColor(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return magenta
	case logrus.InfoLevel:
		return green
	case logrus.WarnLevel:
		return yellow
	case logrus.ErrorLevel:
		return red
	case logrus.FatalLevel, logrus.PanicLevel:
		return bgRed + white
	default:
		return reset
	}
}

// formatValue formats a field value with appropriate styling
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return cyan + "\"" + val + "\"" + dim
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return blue + fmt.Sprintf("%v", val) + dim
	case float32, float64:
		return blue + fmt.Sprintf("%v", val) + dim
	case bool:
		if val {
			return green + "true" + dim
		}
		return red + "false" + dim
	default:
		return yellow + fmt.Sprintf("%v", val) + dim
	}
}

// extractFunctionName extracts the function name from a full path
func extractFunctionName(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		dotIndex := strings.LastIndex(lastPart, ".")
		if dotIndex > 0 {
			return lastPart[dotIndex+1:]
		}
		return lastPart
	}
	return fullPath
}

// extractFileName extracts the filename from a full path
func extractFileName(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}

// isTerminal checks if the output is a terminal
func isTerminal() bool {
	// On Windows, always enable colors for modern terminals (Windows Terminal, VS Code, etc.)
	// The os.ModeCharDevice check doesn't work reliably on Windows
	return true
}

// Init initializes the logger
func Init(level, format string) *logrus.Logger {
	log = logrus.New()

	// Set output
	log.SetOutput(os.Stdout)

	// Enable caller information for better debugging
	log.SetReportCaller(true)

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	// Set formatter
	if format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	} else {
		// Use custom colorful formatter for text output
		log.SetFormatter(&CustomTextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     isTerminal(), // Enable colors for all terminals
		})
	}

	return log
}

// Get returns the logger instance
func Get() *logrus.Logger {
	if log == nil {
		return Init("info", "json")
	}
	return log
}

// SetOutput sets the logger output
func SetOutput(output io.Writer) {
	if log != nil {
		log.SetOutput(output)
	}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Get().Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Get().Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Get().Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Get().Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}

// WithField creates a new entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return Get().WithField(key, value)
}

// WithFields creates a new entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Get().WithFields(fields)
}
