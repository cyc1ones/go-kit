package gorm

import (
	"context"
	"database/sql"

	"github.com/cyc1ones/go-kit/db/transaction"
	"gorm.io/gorm"
)

type Transaction struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Transaction {
	return &Transaction{
		db: db,
	}
}

type txKey struct{}

// DBFromContext implements transaction.Transaction.
func (t *Transaction) DBFromContext(ctx context.Context) any {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return t.db.WithContext(ctx)
}

// Exec implements transaction.Transaction.
func (t *Transaction) Exec(ctx context.Context, fn func(ctx context.Context) error, opts ...*sql.TxOptions) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, txKey{}, tx)
		return fn(ctx)
	})
}

var _ transaction.Transaction = (*Transaction)(nil)
