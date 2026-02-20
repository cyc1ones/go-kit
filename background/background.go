package background

import (
	"context"
	"errors"
	"runtime/debug"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
)

var (
	ErrBackgroundIsRunning = errors.New("background is running")
)

type Task func(ctx context.Context)

// Background is a goroutine manager
type Background struct {
	tasks   []Task
	wg      *sync.WaitGroup
	running bool
	log     *log.Helper
	cancel  context.CancelFunc
}

func New(opts ...Option) *Background {
	s := &Background{
		tasks:   nil,
		wg:      &sync.WaitGroup{},
		running: false,
		log:     newHelper(log.DefaultLogger),
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

// AddTask add a new task to run to background
func (b *Background) AddTask(t Task) error {
	if b.running {
		return ErrBackgroundIsRunning
	}
	b.tasks = append(b.tasks, t)
	return nil
}

func (b *Background) runTask(ctx context.Context, index int) {
	defer b.wg.Done()

	task := b.tasks[index]
	for {
		select {
		case <-ctx.Done():
			return
		default:
			p, stk := noPanic(ctx, task)
			if p != nil {
				b.log.WithContext(ctx).Errorw(
					"msg", "panic during run task",
					"id", index,
					"panic", p,
					"stack", stk,
				)
				continue
			}
			return
		}
	}
}

// Launch run all added tasks in background
func (b *Background) Launch(ctx context.Context) {
	b.running = true

	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	for i := range len(b.tasks) {
		b.wg.Add(1)
		go b.runTask(ctx, i)
	}
	b.log.Infof("[Background] tasks launched: %d", len(b.tasks))
}

// Shutdown cancels the tasks' context
// and waits until all tasks have completed or the given context is canceled.
func (b *Background) Shutdown(ctx context.Context) error {
	b.log.Infof("[Background] tasks stopping")

	b.cancel()

	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// noPanic executes the given task, suppresses any panic it raises,
// and returns the recovered panic value and stack trace.
func noPanic(ctx context.Context, t Task) (p any, stk string) {
	defer func() {
		p = recover()
		if p != nil {
			stk = string(debug.Stack())
		}
	}()

	t(ctx)
	return
}

func newHelper(logger log.Logger) *log.Helper {
	return log.NewHelper(log.With(logger, "module", "background"))
}

type Option func(s *Background)

func WithLogger(logger log.Logger) Option {
	return func(s *Background) {
		s.log = newHelper(logger)
	}
}
