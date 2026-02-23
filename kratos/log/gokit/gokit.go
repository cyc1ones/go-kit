package gokit

import (
	gokitlog "github.com/go-kit/log"
	"github.com/go-kratos/kratos/v2/log"
)

type logger struct {
	base gokitlog.Logger
}

func New(base gokitlog.Logger) log.Logger {
	return &logger{
		base: base,
	}
}

func (l *logger) Log(level log.Level, keyvals ...any) error {
	return l.base.Log(append([]any{"level", level}, keyvals...)...)
}
