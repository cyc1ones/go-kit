package reverseproxy

import (
	"fmt"
	"net/http"
)

var _ http.RoundTripper = (*TransportWrapper)(nil)

type TransportWrapper struct {
	rt http.RoundTripper
}

func (t *TransportWrapper) RoundTrip(r *http.Request) (*http.Response, error) {
	if r == nil {
		return nil, ErrRequestPrevented
	}

	resp, err := t.rt.RoundTrip(r)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}
	return resp, nil
}
