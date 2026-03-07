package reverseproxy

import (
	"context"
	"net/http"
	"strings"
)

// OutgoingRequestHandler should handle outgoing request before it be sent
type OutgoingRequestHandler func(ctx context.Context, req *http.Request) error

// UpstreamResponseHandler should handle response from the upstream before it be sent to client
type UpstreamResponseHandler func(ctx context.Context, resp *http.Response) error

type router struct {
	outgoingRequestHandlers  map[string]OutgoingRequestHandler
	upstreamResponseHandlers map[string]UpstreamResponseHandler
}

func newRouter() *router {
	return &router{
		outgoingRequestHandlers:  make(map[string]OutgoingRequestHandler),
		upstreamResponseHandlers: make(map[string]UpstreamResponseHandler),
	}
}

func (r *router) MatchOutgoingRequestHandler(operation string) OutgoingRequestHandler {
	return r.outgoingRequestHandlers[strings.ToLower(operation)]
}

func (r *router) MatchUpstreamResponseHandler(operation string) UpstreamResponseHandler {
	return r.upstreamResponseHandlers[strings.ToLower(operation)]
}

func (r *router) HandleOutgoingRequest(operation string, h OutgoingRequestHandler) {
	r.outgoingRequestHandlers[strings.ToLower(operation)] = h
}

func (r *router) HandleUpstreamResponse(operation string, handler UpstreamResponseHandler) {
	r.upstreamResponseHandlers[strings.ToLower(operation)] = handler
}
