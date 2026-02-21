package errors

import (
	"github.com/go-kratos/kratos/v2/errors"
)

var (
	ErrInternalServer = errors.InternalServer(
		"INTERNAL_SERVER",
		"内部服务器错误",
	)
)

func Sanitize(err error) error {
	if err == nil {
		return nil
	}
	ke := errors.FromError(err)
	if ke != nil && ke.Reason == errors.UnknownReason {
		return ErrInternalServer.WithCause(err)
	}
	return ke
}
