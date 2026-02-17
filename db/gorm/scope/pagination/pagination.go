package pagination

import (
	"strings"

	"github.com/unkmonster/go-kit/db/gormutil/scope"
	"github.com/unkmonster/go-kit/db/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type options struct {
	stableOrderKey string
}

type Option func(o *options)

// WithStableOrderKey
// key == nil: not apply
// *key == "": use pk as stable order key
// *key == "some_key": use "some_key" as stable order key
func WithStableOrderKey(key string) Option {
	return func(o *options) {
		o.stableOrderKey = key
	}
}

func New(m *query.Pagination, opts ...Option) scope.ScopeFunc {
	options := &options{
		stableOrderKey: "",
	}
	for _, opt := range opts {
		opt(options)
	}

	return func(db *gorm.DB) *gorm.DB {
		if m == nil {
			return db
		}

		// offset
		db = db.Offset(int(m.Offset))

		// limit
		if m.Limit != 0 {
			db = db.Limit(int(m.Limit))
		}

		// sort and order
		if m.SortBy != "" {
			db = db.Order(clause.OrderByColumn{
				Column: clause.Column{Name: m.SortBy},
				Desc:   strings.EqualFold("desc", m.Order),
			})

			// stable order
			if options.stableOrderKey != "" && options.stableOrderKey != m.SortBy {
				db = db.Order(options.stableOrderKey)
			}
		}

		return db
	}
}
