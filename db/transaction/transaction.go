package transaction

import (
	"context"
	"database/sql"
)

type Transaction interface {
	Exec(ctx context.Context, fn func(ctx context.Context) error, opts ...*sql.TxOptions) error
	DBFromContext(ctx context.Context) any
}
