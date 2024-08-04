package generator

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNewFromFunctionWithContext(t *testing.T) {
	ctx := context.Background()
	i := 0
	generate := func() (int, bool) {
		if i < 10 {
			i++
			return i, true
		}

		return 0, false
	}
	items := make([]int, 0)
	for item := range NewFromFunctionWithContext(ctx, generate) {
		items = append(items, item)
	}

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, items)
}

func TestNewFromFunctionWithContextClose(t *testing.T) {
	ctx, closer := context.WithCancel(context.Background())
	defer closer()

	i := 0
	generate := func() (int, bool) {
		i++
		if i == 5 {
			closer()
			time.Sleep(time.Millisecond * 100)
		}

		return i, true
	}

	items := make([]int, 0)
	for item := range NewFromFunctionWithContext(ctx, generate) {
		items = append(items, item)
	}

	assert.Equal(t, []int{1, 2, 3, 4}, items)
}

func TestNewFromMapWithContext(t *testing.T) {
	ctx := context.Background()
	generatorMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
		"seven": 7,
		"eight": 8,
		"nine":  9,
		"ten":   10,
	}
	modifier := func(key string, value int) (string, int) {
		return key + "+one", value + 1
	}
	items := make(map[string]int)
	for item := range NewFromMapWithContext(ctx, generatorMap, modifier) {
		items[item.Key] = item.Value
	}

	assert.Equal(t, map[string]int{
		"one+one":   2,
		"two+one":   3,
		"three+one": 4,
		"four+one":  5,
		"five+one":  6,
		"six+one":   7,
		"seven+one": 8,
		"eight+one": 9,
		"nine+one":  10,
		"ten+one":   11,
	}, items)
}

func TestNewFromMapWithContextClose(t *testing.T) {
	ctx, closer := context.WithCancel(context.Background())
	defer closer()

	generatorMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
		"seven": 7,
		"eight": 8,
		"nine":  9,
		"ten":   10,
	}
	modifier := func(key string, value int) (string, int) {
		return key + "+one", value + 1
	}
	items := make(map[string]int)
	i := 0
	for item := range NewFromMapWithContext(ctx, generatorMap, modifier) {
		i++
		if i == 5 {
			closer()
			time.Sleep(time.Millisecond * 100)
		}
		items[item.Key] = item.Value
	}

	assert.Len(t, items, 6)
	for key, item := range items {
		value, ok := map[string]int{
			"one+one":   2,
			"two+one":   3,
			"three+one": 4,
			"four+one":  5,
			"five+one":  6,
			"six+one":   7,
			"seven+one": 8,
			"eight+one": 9,
			"nine+one":  10,
			"ten+one":   11,
		}[key]
		assert.True(t, ok)
		assert.Equal(t, value, item)
	}
}

func TestNewFromSyncMapWithContext(t *testing.T) {
	ctx := context.Background()
	generatorMap := &sync.Map{}
	generatorMap.Store("one", 1)
	generatorMap.Store("two", 2)
	generatorMap.Store("three", 3)
	generatorMap.Store("four", 4)
	generatorMap.Store("five", 5)
	generatorMap.Store("six", 6)
	generatorMap.Store("seven", 7)
	generatorMap.Store("eight", 8)
	generatorMap.Store("nine", 9)
	generatorMap.Store("ten", 10)
	modifier := func(key string, value int) (string, int) {
		return key + "+one", value + 1
	}
	items := make(map[string]int)
	for item := range NewFromSyncMapWithContext(ctx, generatorMap, modifier) {
		items[item.Key] = item.Value
	}

	assert.Equal(t, map[string]int{
		"one+one":   2,
		"two+one":   3,
		"three+one": 4,
		"four+one":  5,
		"five+one":  6,
		"six+one":   7,
		"seven+one": 8,
		"eight+one": 9,
		"nine+one":  10,
		"ten+one":   11,
	}, items)
}

func TestNewFromSyncMapWithContextClose(t *testing.T) {
	ctx, closer := context.WithCancel(context.Background())
	defer closer()

	generatorMap := &sync.Map{}
	generatorMap.Store("one", 1)
	generatorMap.Store("two", 2)
	generatorMap.Store("three", 3)
	generatorMap.Store("four", 4)
	generatorMap.Store("five", 5)
	generatorMap.Store("six", 6)
	generatorMap.Store("seven", 7)
	generatorMap.Store("eight", 8)
	generatorMap.Store("nine", 9)
	generatorMap.Store("ten", 10)
	modifier := func(key string, value int) (string, int) {
		return key + "+one", value + 1
	}
	items := make(map[string]int)
	i := 0
	for item := range NewFromSyncMapWithContext(ctx, generatorMap, modifier) {
		i++
		if i == 5 {
			closer()
			time.Sleep(time.Millisecond * 100)
		}
		items[item.Key] = item.Value
	}

	assert.Len(t, items, 6)
	for key, item := range items {
		value, ok := map[string]int{
			"one+one":   2,
			"two+one":   3,
			"three+one": 4,
			"four+one":  5,
			"five+one":  6,
			"six+one":   7,
			"seven+one": 8,
			"eight+one": 9,
			"nine+one":  10,
			"ten+one":   11,
		}[key]
		assert.True(t, ok)
		assert.Equal(t, value, item)
	}
}

func TestNewFromMapOnlyValueWithContext(t *testing.T) {
	ctx := context.Background()
	generatorMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
		"seven": 7,
		"eight": 8,
		"nine":  9,
		"ten":   10,
	}
	modifier := func(value int) int {
		return value + 1
	}
	items := make([]int, 0)
	for item := range NewFromMapOnlyValueWithContext(ctx, generatorMap, modifier) {
		items = append(items, item)
	}

	assert.ElementsMatch(t, []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, items)
}

func TestNewFromMapOnlyValueWithContextClose(t *testing.T) {
	ctx, closer := context.WithCancel(context.Background())
	defer closer()

	generatorMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
		"seven": 7,
		"eight": 8,
		"nine":  9,
		"ten":   10,
	}
	modifier := func(value int) int {
		return value + 1
	}
	items := make([]int, 0)
	i := 0
	for item := range NewFromMapOnlyValueWithContext(ctx, generatorMap, modifier) {
		i++
		if i == 5 {
			closer()
			time.Sleep(time.Millisecond * 100)
		}
		items = append(items, item)
	}

	assert.Len(t, items, 6)
	fmt.Println(items)
	for _, item := range items {
		assert.Contains(t, []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, item)
	}
}

func TestNewFromSyncMapOnlyValueWithContext(t *testing.T) {
	ctx := context.Background()
	generatorMap := &sync.Map{}
	generatorMap.Store("one", 1)
	generatorMap.Store("two", 2)
	generatorMap.Store("three", 3)
	generatorMap.Store("four", 4)
	generatorMap.Store("five", 5)
	generatorMap.Store("six", 6)
	generatorMap.Store("seven", 7)
	generatorMap.Store("eight", 8)
	generatorMap.Store("nine", 9)
	generatorMap.Store("ten", 10)
	modifier := func(value int) int {
		return value + 1
	}
	items := make([]int, 0)
	for item := range NewFromSyncMapOnlyValueWithContext(ctx, generatorMap, modifier) {
		items = append(items, item)
	}

	assert.ElementsMatch(t, []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, items)
}

func TestNewFromSyncMapOnlyValueWithContextClose(t *testing.T) {
	ctx, closer := context.WithCancel(context.Background())
	defer closer()

	generatorMap := &sync.Map{}
	generatorMap.Store("one", 1)
	generatorMap.Store("two", 2)
	generatorMap.Store("three", 3)
	generatorMap.Store("four", 4)
	generatorMap.Store("five", 5)
	generatorMap.Store("six", 6)
	generatorMap.Store("seven", 7)
	generatorMap.Store("eight", 8)
	generatorMap.Store("nine", 9)
	generatorMap.Store("ten", 10)
	modifier := func(value int) int {
		return value + 1
	}
	items := make([]int, 0)
	i := 0
	for item := range NewFromSyncMapOnlyValueWithContext(ctx, generatorMap, modifier) {
		i++
		if i == 5 {
			closer()
			time.Sleep(time.Millisecond * 100)
		}
		items = append(items, item)
	}

	assert.Len(t, items, 6)
	for _, item := range items {
		assert.Contains(t, []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, item)
	}
}
