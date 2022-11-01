package parallel

import (
	"context"
	"time"

	"github.com/samsarahq/go/oops"
	"golang.org/x/sync/errgroup"
)

const DefaultParallelism = 100
const DefaultTimeout = 60 * time.Second

type parallelOptions struct {
	parallelism int
	timeout     time.Duration
}

type Option func(*parallelOptions)

func WithMaxConcurrency(parallelism int) Option {
	return func(p *parallelOptions) {
		p.parallelism = parallelism
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(p *parallelOptions) {
		p.timeout = timeout
	}
}

type ForFn[T any] func(ctx context.Context, item T) error

// For performs work for each item in the input slice in parallel.
// It performs work for all items and returns the first encountered error.
// In practice, execution of all subsequent items will be cancelled
// right away because the context is cancelled upon first error.
func For[T any](
	ctx context.Context,
	items []T,
	fn ForFn[T],
	options ...Option,
) error {
	opts := parallelOptions{
		parallelism: DefaultParallelism,
		timeout:     DefaultTimeout,
	}

	for _, o := range options {
		o(&opts)
	}

	ctx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.SetLimit(opts.parallelism)

	for _, item := range items {
		item := item
		errGroup.Go(func() error {
			return oops.Wrapf(fn(ctx, item), "could not perform work for %T item %+v", item, item)
		})
	}

	return oops.Wrapf(errGroup.Wait(), "error executing work")
}

type MapFn[T any, T2 any] func(ctx context.Context, item T) (T2, error)

// Map maps input items to output items in parallel.
// It performs work for all items and returns a results slice and the first encountered error.
// In practice, execution of all subsequent items will be cancelled
// right away because the context is cancelled upon first error.
func Map[T any, T2 any](
	ctx context.Context,
	items []T,
	fn MapFn[T, T2],
	options ...Option,
) ([]T2, error) {
	opts := parallelOptions{
		parallelism: DefaultParallelism,
		timeout:     DefaultTimeout,
	}

	for _, o := range options {
		o(&opts)
	}

	ctx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	errGroup, ctx := errgroup.WithContext(ctx)
	errGroup.SetLimit(opts.parallelism)

	res := make([]T2, len(items))
	for i, item := range items {
		i := i
		item := item
		errGroup.Go(func() error {
			out, err := fn(ctx, item)
			if err != nil {
				return oops.Wrapf(err, "could not compute result for %T item %v", item, item)
			}

			res[i] = out
			return nil
		})
	}

	return res, oops.Wrapf(errGroup.Wait(), "error executing work")
}
