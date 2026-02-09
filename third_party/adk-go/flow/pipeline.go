package flow

import "context"

type Pipeline[I any, A any, B any, C any, D any, O any] struct {
	First  Node[I, A]
	Second Node[A, B]
	Third  Node[B, C]
	Fourth Node[C, D]
	Fifth  Node[D, O]
}

func (p Pipeline[I, A, B, C, D, O]) Execute(ctx context.Context, input I) (O, error) {
	var zero O
	step1, err := p.First.Run(ctx, input)
	if err != nil {
		return zero, err
	}
	step2, err := p.Second.Run(ctx, step1)
	if err != nil {
		return zero, err
	}
	step3, err := p.Third.Run(ctx, step2)
	if err != nil {
		return zero, err
	}
	step4, err := p.Fourth.Run(ctx, step3)
	if err != nil {
		return zero, err
	}
	return p.Fifth.Run(ctx, step4)
}
