//
// Log implemenation for all microservices in the project.
// A logger can be retrieved by calling the Log() function.
//
package logger

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	logger      *logrus.Logger
	logLevel    string
	serviceName string
)

func init() {
	logger = logrus.StandardLogger()
	grpclog.SetLogger(logger)
}

// CreateLogger creates a logger with default settings compatible with StackDriver. Default logLevel is Error.
// However, you can overwrite that by setting your designated log level to the LOG_LEVEL environment variable.
// This function will fail if LOG_LEVEL set to an undefined log level.
func CreateLogger() {
	SetServiceName()
	SetLogLevel()
	setupFormatterStackDriver()
	logger.AddHook(datadogHook{})
}

// Log exports the configured logger ready to use.
func Log() *logrus.Logger {
	return logger
}

// SetServiceName allows overwriting default service name 'hostname'
func SetServiceName(params ...string) {
	if len(params) > 0 {
		serviceName = params[0]
	}

	if serviceName == "" {
		serviceName, _ = os.Hostname()
	}
}

// SetLogLevel allows overwriting default log level 'Error'
func SetLogLevel(params ...string) {
	if len(params) > 0 {
		logLevel = params[0]
	}

	if logLevel == "" {
		logLevel = os.Getenv("LOG_LEVEL")
	}

	logrus.SetLevel(getLogLevel())
}

func setupFormatterStackDriver() {
	logger.Formatter = stackdriver.NewFormatter(
		stackdriver.WithService(serviceName),
	)
}

func getLogLevel() logrus.Level {
	switch logLevel {
	case "": // not set
		return logrus.ErrorLevel
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "error":
		return logrus.ErrorLevel
	case "warn":
		return logrus.WarnLevel
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "trace":
		return logrus.TraceLevel
	}

	panic(fmt.Sprintf("LOG_LEVEL %s is not known", logLevel))
}

// this hook adds datadog tracing metadata to log entry for correlation in datadog
type datadogHook struct {
}

func (d datadogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (d datadogHook) Fire(entry *logrus.Entry) error {
	if span, ok := tracer.SpanFromContext(entry.Context); ok {
		entry.Data["dd"] = map[string]interface{}{
			"trace_id": span.Context().TraceID(),
			"span_id":  span.Context().SpanID(),
		}
	}
	return nil
}
