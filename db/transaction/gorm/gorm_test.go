package gorm

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestShouldCommit(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqldb.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqldb,
		SkipInitializeWithVersion: true,
	}))
	require.NoError(t, err)

	tx := Transaction{db: db}

	err = tx.Exec(context.Background(), func(ctx context.Context) error {
		return nil
	})
	require.NoError(t, err)

	mock.ExpectationsWereMet()
}

func TestShouldRollback(t *testing.T) {
	sqldb, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqldb.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqldb,
		SkipInitializeWithVersion: true,
	}))
	require.NoError(t, err)

	tx := Transaction{db: db}

	err = tx.Exec(context.Background(), func(ctx context.Context) error {
		return errors.New("will rollback")
	})
	require.Error(t, err)

	mock.ExpectationsWereMet()
}

func TestExtractDb(t *testing.T) {
	sqldb, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqldb.Close()

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqldb,
		SkipInitializeWithVersion: true,
	}))
	require.NoError(t, err)

	tx := Transaction{db: db}
	val := tx.DBFromContext(context.Background())
	require.IsType(t, val, &gorm.DB{})
	require.IsType(t, val.(*gorm.DB).Statement.ConnPool, &sql.DB{})

	tx.Exec(context.Background(), func(ctx context.Context) error {
		tx := tx.DBFromContext(ctx)

		require.IsType(t, tx.(*gorm.DB).Statement.ConnPool, &sql.Tx{})
		return nil
	})
}
