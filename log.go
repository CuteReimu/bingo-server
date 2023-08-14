package main

import (
	"github.com/dgraph-io/badger/v3"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"io"
	"log/slog"
	"os"
	"path"
	"time"
)

var log *slog.Logger

func init() {
	writer, err := rotatelogs.New(
		path.Join("log", "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	log = slog.New(slog.NewTextHandler(&logWriter{fileWriter: writer}, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
}

func IsErrKeyNotFound(err error) bool {
	return errors.Is(err, badger.ErrKeyNotFound)
}

type logWriter struct {
	fileWriter io.Writer
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	_, _ = os.Stdout.Write(p)
	return l.fileWriter.Write(p)
}
