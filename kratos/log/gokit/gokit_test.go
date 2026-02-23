package gokit

import (
	"bytes"
	"testing"

	gokitlog "github.com/go-kit/log"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/require"
)

func TestGoKit(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := New(gokitlog.NewLogfmtLogger(buf))
	log := log.NewHelper(logger)
	log.Infof("hello world")

	r := buf.String()
	require.Equal(t, `level=INFO msg="hello world"`+"\n", r)
}
