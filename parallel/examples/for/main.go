package main

import (
	"context"
	"fmt"
	"time"

	"github.com/samsarahq/go/oops"
	"github.com/wit221/go/parallel"
)

func NewPtr[T any](in T) *T {
	return &in
}

type In struct {
	Text string
}

func main() {
	ctx := context.Background()

	items := []*In{
		{"Hi"},
		{"Hello"},
		{"Bye"},
		{"Goodbye"},
	}

	delay := NewPtr(1 * time.Second)

	workFn := func(ctx context.Context, item *In) error {
		select {
		case <-time.After(*delay):
		case <-ctx.Done():
			return oops.Errorf("failed to process item %s", item.Text)
		}

		item.Text = fmt.Sprintf("%s processed", item.Text)
		return nil
	}

	fmt.Println("Run 1 start. Should take 1 second since we have more parallelism available than items.")
	err := parallel.For(ctx, items, workFn)
	fmt.Println(err)
	for _, item := range items {
		fmt.Println(item)
	}

	fmt.Println("Run 2 start. Should take 6 seconds since we have 4 items, 2 parallelism, and each takes 3 seconds")
	delay = NewPtr(3 * time.Second)
	err = parallel.For(ctx, items, workFn, parallel.WithMaxConcurrency(2))
	fmt.Println(err)
	for _, item := range items {
		fmt.Println(item)
	}

	fmt.Println("Run 3 start. Should error because timeout is less than execution time")
	err = parallel.For(ctx, items, workFn, parallel.WithTimeout(2*time.Second))
	fmt.Println(err)
	for _, item := range items {
		fmt.Println(item)
	}
}
