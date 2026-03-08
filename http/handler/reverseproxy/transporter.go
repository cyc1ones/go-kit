package reverseproxy

import (
	"context"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-kratos/kratos/v2/selector"
)

type Transporter struct {
	Upstream        *url.URL
	Operation       string
	IncomingRequest *http.Request

	UpstreamStatusCode int
	Error              error

	requestID string
	done      selector.DoneFunc
	record    *sync.Map
	doneOnce  *sync.Once
}

func NewTransporter() *Transporter {
	return &Transporter{
		record:   new(sync.Map),
		doneOnce: new(sync.Once),
	}
}

func (t *Transporter) Done(ctx context.Context, di selector.DoneInfo) {
	t.doneOnce.Do(func() {
		t.done(ctx, di)
	})
}

func (t *Transporter) RequestID() string {
	return t.requestID
}

type transporterKey struct{}

func NewContext(ctx context.Context, tr *Transporter) context.Context {
	return context.WithValue(ctx, transporterKey{}, tr)
}

func TransporterFromContext(ctx context.Context) (tr *Transporter, ok bool) {
	tr, ok = ctx.Value(transporterKey{}).(*Transporter)
	return
}

func MustTransporterFromContext(ctx context.Context) (tr *Transporter) {
	tr, ok := TransporterFromContext(ctx)
	if !ok {
		panic("missing transporter in context")
	}
	return tr
}
