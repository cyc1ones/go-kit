package reverseproxy

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	testOperation := "login"
	r := newRouter()

	r.HandleOutgoingRequest(testOperation, func(ctx context.Context, req *http.Request) (*http.Request, error) {
		return nil, nil
	})

	r.HandleUpstreamResponse(testOperation, func(ctx context.Context, resp *http.Response) (*http.Response, error) {
		return nil, nil
	})

	or := r.MatchOutgoingRequestHandler(testOperation)
	ur := r.MatchUpstreamResponseHandler(testOperation)

	require.NotEmpty(t, or)
	require.NotEmpty(t, ur)
}
