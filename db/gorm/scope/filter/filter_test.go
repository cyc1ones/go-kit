package filter

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestKv(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})
	require.NoError(t, err)

	kvs := []any{
		"str", "text",
		"num", 2,
	}

	m := make(map[string]any)
	for i := 0; i < len(kvs); i += 2 {
		m[kvs[i].(string)] = kvs[i+1]
	}

	res := db.Table("test").Scopes(
		New(WithKv(m)),
	).Find(nil)

	sql := res.Statement.SQL.String()
	for i := 0; i < len(kvs); i += 2 {
		key := kvs[i].(string)
		require.Contains(t, sql, key)
	}
}
