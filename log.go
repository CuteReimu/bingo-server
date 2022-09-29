package main

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"path"
	"time"
)

var log = &errorEntryWithStack{logrus.WithFields(logrus.Fields{})}

func init() {
	writer, err := rotatelogs.New(
		path.Join("log", "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		logrus.WithError(err).Error("unable to write logs")
		return
	}
	logrus.AddHook(lfshook.NewHook(lfshook.WriterMap{
		logrus.FatalLevel: writer,
		logrus.ErrorLevel: writer,
		logrus.WarnLevel:  writer,
		logrus.InfoLevel:  writer,
		logrus.DebugLevel: writer,
	}, &logrus.TextFormatter{DisableQuote: true}))
	logrus.SetFormatter(&logrus.TextFormatter{DisableQuote: true})
	logrus.SetLevel(logrus.DebugLevel)
}

type errorEntryWithStack struct {
	*logrus.Entry
}

func (e *errorEntryWithStack) WithError(err error) *logrus.Entry {
	return e.Entry.WithError(fmt.Errorf("%+v", err))
}
