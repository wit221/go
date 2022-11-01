# What

Simple module for one-off fan-out workloads. Akin to C's OpenMP parallelism pragmas.

# Install
`go get github.com/wit221/go/parallel`

# Examples

See `examples/**/main.go`. Basic examples:

## For:

```go
func main() {
    ctx := context.Background()

    items := []*In{
        {"Hi"},
        {"Hello"},
        {"Bye"},
        {"Goodbye"},
    }

    workFn := func(ctx context.Context, item *In) error {
        select {
        case <-time.After(time.Second):
        case <-ctx.Done():
            return oops.Errorf("failed to process item %s", item.Text)
        }

        item.Text = fmt.Sprintf("%s processed", item.Text)
        return nil
    }
    err := parallel.For(ctx, items, workFn)
    // ...
}
```

## Map:

```go
func main() {
	ctx := context.Background()

	items := []*In{
		{"Hi"},
		{"Hello"},
		{"Bye"},
		{"Goodbye"},
	}

	delay := NewPtr(1 * time.Second)

	workFn := func(ctx context.Context, item *In) (*Out, error) {
		select {
		case <-time.After(*delay):
		case <-ctx.Done():
			return nil, oops.Errorf("failed to process item %s", item.Text)
		}

		return &Out{ProcessedText: fmt.Sprintf("%s processed", item.Text)}, nil
	}

	res, err := parallel.Map(ctx, items, workFn)
    // ...
}
```
