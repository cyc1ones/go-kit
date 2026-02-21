package reverseproxy

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-kratos/kratos/v2/selector"
)

type Transporter struct {
	Upstream        *url.URL
	Operation       string
	IncomingRequest *http.Request
	Done            selector.DoneFunc
}

type transporterKey struct{}

func NewContext(ctx context.Context, tr *Transporter) context.Context {
	return context.WithValue(ctx, transporterKey{}, tr)
}

func TransporterFromContext(ctx context.Context) (tr *Transporter, ok bool) {
	tr, ok = ctx.Value(transporterKey{}).(*Transporter)
	return
}
