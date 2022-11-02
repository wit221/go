# What

Simple module for throttling execution by string key in-memory.
Example use case: You want to notify a user whenever an event of a certain type happens.
But, if that event fires multiple times in a succession, you don't want to send them multiple
notifications. Instead, you want to send one notification for the first event, and then
another one for multiple remaining events in a provided time span.
The throttler works on multiple keys and throttles separately for each key.
It does grab a mutex so you may encounter limitations in very high contention environments.
It obviously does not hold it for asynchronous callback invocations - only for synchronous updates.

# Install

`go get github.com/wit221/go/throttle`

# Examples

See `examples/**/main.go`. Basic examples:

## Throttle:

```go
callback := func(ctx context.Context, key string, items []*Pair, instantCallback bool) error {
    values := make([]string, 0, len(items))
    for _, item := range items {
        values = append(values, item.value)
    }

    fmt.Printf("key[%v] values[%s]\n", key, strings.Join(values, ", "))
    return nil
}

ctx := context.Background()
throttler := throttle.New(callback, 10*time.Second)
reader := bufio.NewReader(os.Stdin)
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        case <-time.After(5 * time.Second):
            fmt.Printf("num_goroutines [%v]\n", runtime.NumGoroutine())
        }
    }
}()
for {
    select {
    case <-ctx.Done():
        return
    default:
        text, err := reader.ReadString('\n')
        if err != nil {
            return
        }
        text = strings.Replace(text, "\n", "", -1)
        parts := strings.Split(text, ",")
        go func() {
            pair := &Pair{key: parts[0], value: parts[1]}
            throttler.Submit(ctx, pair.key, pair)
        }()
    }
}
```
