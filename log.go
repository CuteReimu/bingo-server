package main

import (
	"github.com/davyxu/golog"
	"github.com/dgraph-io/badger/v3"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"io"
	"os"
	"path"
	"time"
)

var log = golog.New("bingo")

func init() {
	log.SetLevel(golog.Level_Debug)
	writer, err := rotatelogs.New(
		path.Join("log", "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		panic(err)
	}
	err = golog.SetOutput("bingo", &logWriter{fileWriter: writer})
	if err != nil {
		panic(err)
	}
}

func IsErrKeyNotFound(err error) bool {
	return err == badger.ErrKeyNotFound || errors.Unwrap(err) == badger.ErrKeyNotFound
}

type logWriter struct {
	fileWriter io.Writer
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	_, _ = os.Stdout.Write(p)
	return l.fileWriter.Write(p)
}
