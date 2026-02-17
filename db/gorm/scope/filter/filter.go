package filter

import (
	"fmt"

	"github.com/unkmonster/go-kit/db/gormutil/scope"
	"gorm.io/gorm"
)

type options struct {
	Kv map[string]any
}

type Option func(o *options)

func WithKv(kv map[string]any) Option {
	return func(o *options) {
		o.Kv = kv
	}
}

func New(opts ...Option) scope.ScopeFunc {
	options := new(options)
	for _, opt := range opts {
		opt(options)
	}

	return func(db *gorm.DB) *gorm.DB {
		for k, v := range options.Kv {
			db = db.Where(fmt.Sprintf("%s=?", k), v)
		}
		return db
	}
}
