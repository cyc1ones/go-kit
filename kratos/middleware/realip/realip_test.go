package realip

import (
	"context"
	"fmt"
	nethttp "net/http"
	"testing"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/require"
)

type transporter struct {
	r *http.Request
}

// PathTemplate implements [http.Transporter].
func (t *transporter) PathTemplate() string {
	panic("unimplemented")
}

// Request implements [http.Transporter].
func (t *transporter) Request() *http.Request {
	return t.r
}

// Endpoint implements [transport.Transporter].
func (t *transporter) Endpoint() string {
	panic("unimplemented")
}

// Kind implements [transport.Transporter].
func (t *transporter) Kind() transport.Kind {
	panic("unimplemented")
}

// Operation implements [transport.Transporter].
func (t *transporter) Operation() string {
	panic("unimplemented")
}

// ReplyHeader implements [transport.Transporter].
func (t *transporter) ReplyHeader() transport.Header {
	panic("unimplemented")
}

// RequestHeader implements [transport.Transporter].
func (t *transporter) RequestHeader() transport.Header {
	panic("unimplemented")
}

var _ http.Transporter = (*transporter)(nil)

func TestParseProxy(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "192.168.0.1",
			want:  "192.168.0.1/32",
		},
		{
			input: "198.168.0.0/16",
			want:  "198.168.0.0/16",
		},
		{
			input: "xxxxx",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cidr, err := parseProxy(tt.input)
			if tt.want == "" {
				require.Error(t, err)
				return
			}
			require.Equal(t, tt.want, cidr.String())
		})
	}
}

func TestParseXFF(t *testing.T) {
	tests := []struct {
		xff       string
		recursive bool
		proxies   []string
		want      string
	}{
		// 递归模式，8.8.8.8 可信，结果为 1.1.1.1
		{

			xff:       "1.1.1.1,8.8.8.8",
			recursive: true,
			proxies:   []string{"8.8.8.8"},
			want:      "1.1.1.1",
		},

		// 非递归，总是返回最后一个 IP, 即使它是受信任的代理
		{

			xff:       "1.1.1.1,8.8.8.8",
			recursive: false,
			proxies:   []string{"8.8.8.8"},
			want:      "8.8.8.8",
		},

		// 递归模式，但所有地址都受信任，返回最左的 IP
		{

			xff:       "1.1.1.1,2.2.2.2,3.3.3.3,5.5.5.5",
			recursive: true,
			proxies:   []string{"0.0.0.0/0"},
			want:      "1.1.1.1",
		},
		// 由于 5.5.5.5 是受信任的代理，它转发的 IP 尽管无效也会作为结果
		{

			xff:       "1.1.1.1,2.2.2.2,invalid,5.5.5.5",
			recursive: true,
			proxies:   []string{"5.5.5.5", "2.2.2.2"},
			want:      "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.xff, func(t *testing.T) {
			r := &RealIP{
				recursive: tt.recursive,
			}
			WithProxies(tt.proxies)(r)
			result := r.getRealIPFromXFF(tt.xff)
			require.Equal(t, tt.want, result)
		})
	}
}

func TestRealIP(t *testing.T) {
	xffHeader := DefaultXFFHeader

	tests := []struct {
		xff        string
		remoteAddr string
		recursive  bool
		proxies    []string
		want       string
	}{
		// 递归模式，8.8.8.8 可信，结果为 1.1.1.1
		{
			remoteAddr: "9.9.9.9:999",
			recursive:  true,
			proxies:    []string{"8.8.8.8", "9.9.9.9"},
			xff:        "1.1.1.1,8.8.8.8",
			want:       "1.1.1.1",
		},

		// 非递归，总是返回最后一个 IP, 即使它是受信任的代理
		{

			xff:        "1.1.1.1,8.8.8.8",
			remoteAddr: "9.9.9.9:999",

			recursive: false,
			proxies:   []string{"8.8.8.8", "9.9.9.9"},

			want: "8.8.8.8",
		},

		// 递归模式，但所有地址都受信任，返回最左的 IP
		{

			xff:        "1.1.1.1,2.2.2.2,3.3.3.3,5.5.5.5",
			remoteAddr: "9.9.9.9:999",
			recursive:  true,
			proxies:    []string{"0.0.0.0/0"},

			want: "1.1.1.1",
		},
		// 由于 5.5.5.5 是受信任的代理，它转发的 IP 尽管无效也会作为结果
		{

			xff:        "1.1.1.1,2.2.2.2,invalid,5.5.5.5",
			remoteAddr: "9.9.9.9:999",
			recursive:  true,
			proxies:    []string{"5.5.5.5", "2.2.2.2", "9.9.9.9"},

			want: "invalid",
		},
		// remoteAddr 不可信，结果为 remoteAddr
		{
			xff:        "1.1.1.1",
			remoteAddr: "9.9.9.9:999",
			recursive:  false,
			proxies:    nil,
			want:       "9.9.9.9",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.xff, tt.recursive, tt.proxies), func(t *testing.T) {
			m := New(
				WithProxies(tt.proxies),
				WithRecursive(tt.recursive),
			).Server()

			var next middleware.Handler = func(ctx context.Context, req any) (any, error) {
				addr, ok := FromContext(ctx)
				if tt.want == "" {
					require.False(t, ok)
					return nil, nil
				}
				require.True(t, ok)
				require.Equal(t, tt.want, addr)
				return nil, nil
			}

			ctx := context.Background()
			ctx = transport.NewServerContext(ctx, &transporter{
				r: &nethttp.Request{
					Header: nethttp.Header{
						xffHeader: []string{tt.xff},
					},
					RemoteAddr: tt.remoteAddr,
				},
			})
			_, ok := transport.FromServerContext(ctx)
			require.True(t, ok)

			h := m(next)
			_, err := h(ctx, nil)
			require.NoError(t, err)
		})
	}
}
