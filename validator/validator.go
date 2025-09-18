package validator

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

func Validate(ctx context.Context, data any, handler func(context.Context, error) error) (err error) {
	switch v := data.(type) {
	case interface{ ValidateAll() error }:
		err = v.ValidateAll()
	case interface{ Validate(all bool) error }:
		err = v.Validate(true)
	case interface{ Validate() error }:
		err = v.Validate()
	case proto.Message:
		err = protovalidate.Validate(v)
	}
	if err != nil && handler != nil {
		err = handler(ctx, err)
	}
	return
}
