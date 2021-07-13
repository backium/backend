package core

import (
	"context"
)

type Uploader interface {
	Upload(context.Context, string) (string, error)
}
