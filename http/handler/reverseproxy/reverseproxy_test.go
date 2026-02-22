package reverseproxy

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFixCookieDomain(t *testing.T) {
	tests := []struct {
		raw  string
		from string
		to   string

		want string
	}{
		{
			raw:  "raw.com",
			from: "raw.com",
			to:   "to.com",
			want: "to.com",
		},
		{
			raw:  "raw.com",
			from: "RAW.CoM",
			to:   "To.CoM",
			want: "to.com",
		},
		{
			raw:  "www.Raw.com",
			from: "RAW.CoM",
			to:   "To.CoM",
			want: "www.to.com",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%+v", tt), func(t *testing.T) {
			r := &http.Response{
				Header: map[string][]string{
					"Set-Cookie": {
						fmt.Sprintf("key=value; Domain=%s", tt.raw),
					},
				},
			}
			err := fixCookieDomain(r, tt.from, tt.to)
			require.NoError(t, err)
			require.Contains(t, r.Header.Get("Set-Cookie"), tt.want)
			require.NotContains(t, r.Header.Get("Set-Cookie"), tt.from)
		})
	}
}

func TestBuildNodeAndUpstream(t *testing.T) {
	tests := []struct {
		upstream string
	}{
		{
			upstream: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%+v", tt), func(t *testing.T) {
			nodes := buildNodes([]string{tt.upstream})
			node := nodes[0]

			upstream, err := buildUpstream(node)
			require.NoError(t, err)
			require.True(t, strings.EqualFold(tt.upstream, upstream.String()))
		})
	}
}

func TestTransportWrapper(t *testing.T) {
	tr := transportWrapper{rt: http.DefaultTransport}
	_, err := tr.RoundTrip(nil)
	require.ErrorIs(t, ErrRequestPrevented, err)
}
