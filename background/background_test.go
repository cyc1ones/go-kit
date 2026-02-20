package background

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBackground(t *testing.T) {
	bg := New()

	var total int32 = 5

	launched := &atomic.Int32{}
	closed := &atomic.Int32{}

	for range total {
		bg.AddTask(func(ctx context.Context) {
			launched.Add(1)

			<-ctx.Done()

			closed.Add(1)
		})
	}

	ctx := context.Background()
	bg.Launch(ctx)

	time.Sleep(1 * time.Second)
	// 启动后，cancel 应该被赋值
	require.NotEmpty(t, bg.cancel)

	err := bg.Shutdown(ctx)
	require.NoError(t, err)

	// 全部启动以及全部退出
	require.Equal(t, total, launched.Load())
	require.Equal(t, total, closed.Load())
}

func TestAutoRecover(t *testing.T) {
	p := "ppppp"
	bg := New()

	count := 0
	triggered := false
	bg.AddTask(func(ctx context.Context) {
		count++
		if !triggered {
			triggered = true
			panic(p)
		}
	})
	bg.wg.Add(1)

	bg.runTask(context.Background(), 0)
	require.Equal(t, 2, count)
}
