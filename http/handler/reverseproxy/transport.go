package reverseproxy

import (
	"fmt"
	"net/http"
)

var _ http.RoundTripper = (*transportWrapper)(nil)

type transportWrapper struct {
	rt http.RoundTripper
}

func (t *transportWrapper) RoundTrip(r *http.Request) (*http.Response, error) {
	tr := MustTransporterFromContext(r.Context())

	if err := tr.Error; err != nil {
		return nil, fmt.Errorf("request prevented: %w", err)
	}
	if r == nil {
		return nil, ErrRequestPrevented
	}

	resp, err := t.rt.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}
	return resp, nil
}
