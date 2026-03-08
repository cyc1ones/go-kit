package reverseproxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/node/ewma"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/uuid"
)

var debug = true
var Debug = true

var (
	// ErrRequestPrevented returns when RoundTrip if request is nil
	ErrRequestPrevented = errors.New("request prevented")
)

// ErrorHandlerFunc should handler a error during request
type ErrorHandlerFunc func(rw http.ResponseWriter, r *http.Request, err error)

// MakeOperationFunc should make operation from request
type MakeOperationFunc func(ctx context.Context, r *http.Request) string

type Option func(r *ReverseProxy)

func WithTransport(t http.RoundTripper) Option {
	return func(r *ReverseProxy) {
		r.proxy.Transport = &transportWrapper{t}
	}
}

func WithSelector(selector selector.Selector) Option {
	return func(r *ReverseProxy) {
		r.selector = selector
	}
}

func WithLogger(logger log.Logger) Option {
	return func(r *ReverseProxy) {
		r.log = buildHelper(logger)
	}
}

func WithFixCookieDomain(b bool) Option {
	return func(r *ReverseProxy) {
		r.fixCookieDomain = b
	}
}

func WithFixRequestHeader(b bool) Option {
	return func(r *ReverseProxy) {
		r.fixRequestHeader = b
	}
}

func WithHandleError(f ErrorHandlerFunc) Option {
	return func(r *ReverseProxy) {
		r.handleError = f
	}
}

func WithMakeOperation(f MakeOperationFunc) Option {
	return func(r *ReverseProxy) {
		r.makeOperation = f
	}
}

func WithOutgoingRequestMiddlewares(m ...OutgoingRequestMiddleware) Option {
	return func(r *ReverseProxy) {
		r.outgoingRequestChain = OutgoingRequestChain(m...)
	}
}

func WithUpstreamResponseMiddleware(m ...UpstreamResponseMiddleware) Option {
	return func(r *ReverseProxy) {
		r.upstreamResponseChain = UpstreamResponseChain(m...)
	}
}

func WithRouter(rt Router) Option {
	return func(r *ReverseProxy) {
		r.router = rt
	}
}

type ReverseProxy struct {
	proxy  *httputil.ReverseProxy
	router Router

	log *log.Helper

	selector selector.Selector

	handleError   ErrorHandlerFunc
	makeOperation MakeOperationFunc

	fixCookieDomain  bool
	fixRequestHeader bool

	outgoingRequestChain  OutgoingRequestMiddleware
	upstreamResponseChain UpstreamResponseMiddleware
}

func New(opts ...Option) *ReverseProxy {
	handler := &ReverseProxy{
		log: buildHelper(log.DefaultLogger),
		selector: (&selector.DefaultBuilder{
			Node: &ewma.Builder{
				ErrHandler: func(err error) (isErr bool) {
					return errors.Is(err, io.EOF)
				},
			},
			Balancer: &wrr.Builder{},
		}).Build(),
		handleError:   kratoshttp.DefaultErrorEncoder,
		makeOperation: defaultMakeOperation,

		fixCookieDomain:  true,
		fixRequestHeader: true,

		router: NewRouter(),
	}
	proxy := &httputil.ReverseProxy{
		Transport:      http.DefaultTransport,
		Rewrite:        handler.rewrite,
		ModifyResponse: handler.modifyResponse,
		ErrorHandler:   handler.errorHandler,
	}
	handler.proxy = proxy

	for _, opt := range opts {
		opt(handler)
	}
	return handler
}

// rewrite will be registered to httputil.ReverseProxy
func (rp *ReverseProxy) rewrite(pr *httputil.ProxyRequest) {
	ctx := pr.Out.Context()

	// upstream, operation
	tr := MustTransporterFromContext(ctx)

	var (
		upstream  = tr.Upstream
		operation = tr.Operation
	)

	pr.SetURL(upstream)
	pr.Out.URL.RawQuery = pr.In.URL.RawQuery
	pr.Out.Host = upstream.Host

	if rp.fixRequestHeader {
		origin := upstream.Scheme + "://" + upstream.Host
		pr.Out.Header.Set("Origin", origin)
		pr.Out.Header.Set("Referer", origin+"/")
	}

	// match and call handler
	handler := rp.router.MatchOutgoingRequestHandler(operation)
	if handler == nil {
		handler = func(ctx context.Context, req *http.Request) error {
			return nil
		}
	}

	if rp.outgoingRequestChain != nil {
		handler = rp.outgoingRequestChain(handler)
	}
	err := handler(ctx, pr.Out)
	if err != nil {
		// 这里通过 transportWrapper 传播错误至 errorHandler 处理
		// 如果在 transportWrapper 执行 RoundTrip 时发现 Error 已被设置，会直接返回这个错误
		tr.Error = fmt.Errorf("handle outgoing request: %w", err)
		return
	}

	if Debug {
		d, err := httputil.DumpRequest(pr.Out, true)
		if err != nil {
			rp.log.Warnf("failed to dump outgoing request: %v", err)
		} else {
			rp.log.WithContext(ctx).Debugw(
				"request to be sent", string(d),
			)
		}
	}
}

// modifyResponse will be registered to httputil.ReverseProxy
//
// resp is response from upstream
func (rp *ReverseProxy) modifyResponse(resp *http.Response) error {
	ctx := resp.Request.Context()

	tr := MustTransporterFromContext(ctx)

	tr.UpstreamStatusCode = resp.StatusCode

	if Debug {
		d, err := httputil.DumpResponse(resp, true)
		if err != nil {
			rp.log.WithContext(ctx).Warnf("failed to dump response from upstream: %v", err)
		} else {
			rp.log.WithContext(ctx).Debugw(
				"response from upstream", string(d),
			)
		}
	}

	// fix cookie domain
	if rp.fixCookieDomain {
		host := tr.IncomingRequest.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		} else {
			rp.log.WithContext(ctx).Warnw(
				"msg", "failed to split host port",
				"reason", err,
				"host_header", host,
			)
		}
		if err := fixCookieDomain(resp, "", host); err != nil {
			rp.log.WithContext(ctx).Warnf("failed to fix cookie domain: %v", err)
		}
	}

	// match and call handler
	handler := rp.router.MatchUpstreamResponseHandler(tr.Operation)
	if handler == nil {
		handler = func(ctx context.Context, resp *http.Response) error {
			return nil
		}
	}
	if rp.upstreamResponseChain != nil {
		handler = rp.upstreamResponseChain(handler)
	}
	err := handler(ctx, resp)
	if err != nil {
		return fmt.Errorf("handle upstream response: %w", err)
	}

	if Debug {
		d, err := httputil.DumpResponse(resp, true)
		if err != nil {
			rp.log.WithContext(ctx).Warnf("failed to dump response to be sent: %v", err)
		} else {
			rp.log.WithContext(ctx).Debugw(
				"response to be sent", string(d),
			)
		}
	}

	tr.Done(ctx, selector.DoneInfo{
		BytesSent:     true,
		BytesReceived: true,
	})
	return nil
}

// errorHandler will be register to httputil.ReverseProxy
//
// it calls and pass error to 'done' of selected node and calls handleError
func (rp *ReverseProxy) errorHandler(rw http.ResponseWriter, r *http.Request, err error) {
	ctx := r.Context()
	if tr, ok := TransporterFromContext(ctx); ok {
		tr.Done(ctx, selector.DoneInfo{
			Err: err,
		})
		tr.Error = err
	}
	rp.handleError(rw, r, err)
}

func (rp *ReverseProxy) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// select a node and build it as upstream
	node, done, err := rp.selector.Select(ctx)
	if err != nil {
		rp.errorHandler(rw, r, fmt.Errorf("select node: %w", err))
		return
	}
	upstream, err := buildUpstream(node)
	if err != nil {
		rp.errorHandler(rw, r, fmt.Errorf("build upstream: %w", err))
		return
	}

	// make operation for current request
	operation := rp.makeOperation(ctx, r)

	// inject the Transporter of current request into context
	tr, _ := TransporterFromContext(ctx)
	if tr == nil {
		tr = NewTransporter()
	}

	tr.Upstream = upstream
	tr.Operation = operation
	tr.IncomingRequest = r
	tr.done = done
	tr.requestID = uuid.NewString()
	ctx = NewContext(ctx, tr)

	// debug header
	if debug {
		rw.Header().Set("X-Upstream", upstream.String())
	}

	rp.proxy.ServeHTTP(rw, r.WithContext(ctx))
}

func (rp *ReverseProxy) SetUpstreams(upstreams []string) {
	rp.selector.Apply(buildNodes(upstreams))
}

func buildHelper(logger log.Logger) *log.Helper {
	return log.NewHelper(log.With(logger, "component", "handler/reverseproxy"))
}

// buildUpstream builds a URL from Node
func buildUpstream(node selector.Node) (*url.URL, error) {
	return url.Parse(node.Address())
}

func buildNodes(upstreams []string) []selector.Node {
	nodes := make([]selector.Node, len(upstreams))
	for i, up := range upstreams {
		nodes[i] = selector.NewNode("", up, nil)
	}
	return nodes
}

// fixCookieDomain replaces domain in cookies from 'from' to 'to'
// if 'from' is empty, use outgoing host as default value
func fixCookieDomain(r *http.Response, from, to string) error {
	if to == "" {
		return errors.New("'to' must not be empty")
	}
	if from == "" {
		from = r.Request.URL.Hostname()
	}

	from = strings.ToLower(from)
	to = strings.ToLower(to)

	cookies := r.Cookies()
	r.Header.Del("Set-Cookie")
	for _, c := range cookies {
		// ignore host-only cookie
		if c.Domain == "" {
			r.Header.Add("Set-Cookie", c.String())
			continue
		}

		domain := strings.ToLower(c.Domain)
		if before, ok := strings.CutSuffix(domain, from); ok {
			c.Domain = before + to
		}
		r.Header.Add("Set-Cookie", c.String())
	}
	return nil
}

// defaultMakeOperation extracts URL.Path from given request as operation
func defaultMakeOperation(ctx context.Context, r *http.Request) string {
	return r.URL.Path
}
