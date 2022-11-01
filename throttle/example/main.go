package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/wit221/go/throttle"
)

type Pair struct {
	key   string
	value string
}

func main() {
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
}
