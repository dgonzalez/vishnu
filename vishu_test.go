package main

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createInstance(closingTimeout time.Duration) *Vishu {
	vishu := New(func(stats map[string]interface{}) int {
		return stats["target"].(int)
	}, closingTimeout)
	vishu.Add("test1")
	vishu.Add("test2")
	return vishu
}

func TestHappyPath(t *testing.T) {
	vishu := createInstance(time.Millisecond)
	vishu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		if ctx.Target.(string) == "test1" {
			ctx.Stats["target"] = 1000
		} else {
			ctx.Stats["target"] = 0
		}
		return ctx.Stats, nil
	})

	vishu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		assert.Equal(t, ctx.Target.(string), "test1", "best endpoint selected")
		return ctx.Stats, nil
	})
}

func TestNilEndpoint(t *testing.T) {
	vishu := New(nil, time.Millisecond)
	err := vishu.Add(nil)
	assert.NotNil(t, err)
}

func TestCircuitOpen(t *testing.T) {
	vishu := createInstance(time.Millisecond)
	vishu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		return ctx.Stats, errors.New("Something happened")
	})
	assert.Equal(t, Open, int(vishu.endpoints[0].circuitStatus))
}
