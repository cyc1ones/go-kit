package pagination

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/unkmonster/go-kit/db/query"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   int64
	Name string
}

func TestPagination(t *testing.T) {
	tests := []struct {
		p              *query.Pagination
		stableOrderKey string
		expectSql      string
	}{
		{
			p: &query.Pagination{
				Offset: 0,
				Limit:  20,
				Order:  "desc",
				SortBy: "name",
			},
			stableOrderKey: "id",
			expectSql:      "SELECT * FROM `users` ORDER BY `name` DESC,id LIMIT 20",
		},
		{
			p: &query.Pagination{
				Offset: 15,
				Limit:  0,
				Order:  "desc",
				SortBy: "name",
			},
			stableOrderKey: "id",
			expectSql:      "SELECT * FROM `users` ORDER BY `name` DESC,id LIMIT -1 OFFSET 15",
		},
		{
			p:              nil,
			stableOrderKey: "",
			expectSql:      "SELECT * FROM `users`",
		},
	}

	for i, tt := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
				DryRun: true,
			})
			require.NoError(t, err)

			dst := make([]*User, 0)
			res := db.Model(&User{}).Scopes(
				New(tt.p, WithStableOrderKey(tt.stableOrderKey)),
			).Find(dst)

			require.Equal(t, tt.expectSql, res.Statement.SQL.String())
		})
	}

}
