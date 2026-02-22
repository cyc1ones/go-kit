package reverseproxy

type OutgoingRequestMiddleware func(h OutgoingRequestHandler) OutgoingRequestHandler
type UpstreamResponseMiddleware func(h UpstreamResponseHandler) UpstreamResponseHandler

func OutgoingRequestChain(middlewares ...OutgoingRequestMiddleware) OutgoingRequestMiddleware {
	return func(h OutgoingRequestHandler) OutgoingRequestHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}
		return h
	}
}

func UpstreamResponseChain(middlewares ...UpstreamResponseMiddleware) UpstreamResponseMiddleware {
	return func(h UpstreamResponseHandler) UpstreamResponseHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}
		return h
	}
}
