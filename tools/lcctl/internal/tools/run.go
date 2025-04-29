package tools

import "context"

type Runner interface {
	Run(ctx context.Context, args ...any)
}
