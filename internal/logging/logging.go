package logging

import (
	"context"
	"os"
	"strings"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// Logger is a custom logger type embedding logrus.Logger
type Logger struct {
	*logrus.Entry
}

func (l *Logger) SetDebug() {
	l.Logger.SetLevel(logrus.DebugLevel)
	l.Logger.SetReportCaller(true)
}

func newLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(
		&nested.Formatter{
			HideKeys:      true,
			ShowFullLevel: true,
			FieldsOrder:   []string{"MODULE", "METHOD", "URL"},
		},
	)

	logger.SetOutput(os.Stdout)

	return logger
}

func NewLogger() *Logger {
	logger := newLogger()

	return &Logger{
		logger.WithContext(context.TODO()),
	}
}

func NewLoggerWithModule(module string) *Logger {
	logger := newLogger()

	return &Logger{
		logger.WithField("MODULE", strings.ToUpper(module)),
	}
}
