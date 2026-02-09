package flow

import "context"

// Node defines a typed execution node in an ADK flow.
type Node[I any, O any] interface {
	Name() string
	Run(ctx context.Context, input I) (O, error)
}
