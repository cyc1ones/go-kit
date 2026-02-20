package realip

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"
)

const (
	DefaultXFFHeader = "X-Forwarded-For"
)

type RealIP struct {
	trustedProxies []*net.IPNet
	ipHeaders      []string
	recursive      bool
	log            *log.Helper
}

func New(opts ...Option) *RealIP {
	r := &RealIP{
		trustedProxies: nil,
		ipHeaders:      []string{DefaultXFFHeader},
		recursive:      true,
		log:            newHelper(log.DefaultLogger),
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (o *RealIP) Server() middleware.Middleware {
	return func(h middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			request, ok := http.RequestFromServerContext(ctx)
			if !ok {
				o.log.WithContext(ctx).Warnf("missing HTTP request in context")
				return h(ctx, req)
			}

			addr, _, err := net.SplitHostPort(request.RemoteAddr)
			if err != nil {
				o.log.WithContext(ctx).Warnf("failed to parse remoteAddr: %v", err)
				return h(ctx, req)
			}

			if o.isTrustedProxy(net.ParseIP(addr)) {
				for _, key := range o.ipHeaders {
					value := request.Header.Get(key)
					tempAddr := o.getRealIPFromXFF(value)
					if tempAddr != "" {
						addr = tempAddr
						break
					}
				}
			}

			ctx = NewContext(ctx, addr)
			return h(ctx, req)
		}
	}
}

func (o *RealIP) isTrustedProxy(ip net.IP) bool {
	if ip == nil {
		return false
	}

	for _, proxy := range o.trustedProxies {
		if proxy.Contains(ip) {
			return true
		}
	}
	return false
}

// getRealIPFromXFF parse xff header and return real IP
func (o *RealIP) getRealIPFromXFF(trustedValue string) string {
	if trustedValue == "" {
		return ""
	}

	addresses := strings.Split(strings.TrimSpace(trustedValue), ",")
	n := len(addresses)

	// 非递归模式：直接返回 xff 中的最后一个地址作为客户端 IP
	if n > 0 && !o.recursive {
		return strings.TrimSpace(addresses[n-1])
	}

	// 递归模式：逆序依次验证 IP 是否为可信代理，返回遇到的第一个不受信任的地址
	var last string
	for i := n - 1; i >= 0; i-- {
		last = strings.TrimSpace(addresses[i])
		if !o.isTrustedProxy(net.ParseIP(last)) {
			break
		}
	}
	return last
}

type realIPKey struct{}

func NewContext(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, realIPKey{}, addr)
}

func FromContext(ctx context.Context) (addr string, ok bool) {
	addr, ok = ctx.Value(realIPKey{}).(string)
	return
}

type Option func(o *RealIP)

// WithProxies 支持 IP/CIDR
func WithProxies(proxies []string) Option {
	cidrs := make([]*net.IPNet, 0, len(proxies))
	for _, proxy := range proxies {
		cidr, err := parseProxy(proxy)
		if err != nil {
			panic(fmt.Sprintf("invalid proxy %q: %v", proxy, err))
		}
		cidrs = append(cidrs, cidr)
	}

	return func(o *RealIP) {
		o.trustedProxies = cidrs
	}
}

func WithIPHeaders(headers ...string) Option {
	return func(o *RealIP) {
		o.ipHeaders = append(o.ipHeaders, headers...)
	}
}

func WithRecursive(r bool) Option {
	return func(o *RealIP) {
		o.recursive = r
	}
}

func WithLogger(logger log.Logger) Option {
	return func(o *RealIP) {
		o.log = newHelper(logger)
	}
}

func newHelper(logger log.Logger) *log.Helper {
	return log.NewHelper(log.With(logger, "module", "middleware/realip"))
}

func parseProxy(proxy string) (*net.IPNet, error) {
	if !strings.Contains(proxy, "/") {
		ip := net.ParseIP(proxy)
		if ip.To4() != nil {
			proxy += "/32"
		} else {
			proxy += "/128"
		}
	}

	_, cidr, err := net.ParseCIDR(proxy)
	return cidr, err
}
