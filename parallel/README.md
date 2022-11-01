# What

Simple module for one-off fan-out workloads. Akin to C's OpenMP parallelism pragmas.

# Examples

See `example/\*\*/main.go`. Basic examples:

## For:

```
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
...
```

## Map:

```

```
