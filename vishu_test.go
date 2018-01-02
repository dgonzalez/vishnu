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
	assert.Equal(t, Closed, int(vishu.endpoints[1].circuitStatus))
}

func TestHalfOpenAfterTimeout(t *testing.T) {
	vishu := createInstance(time.Millisecond)

	// Break one endpoint
	vishu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		ctx.Stats["target"] = 1000
		return ctx.Stats, errors.New("Something happened")
	})
	// Wait for the circuit to be half-open
	time.Sleep(time.Millisecond * 3)

	// Check that was switched half-open
	assert.Equal(t, HalfOpen, int(vishu.endpoints[0].circuitStatus))
	assert.Equal(t, Closed, int(vishu.endpoints[1].circuitStatus))

	// Trick to force choosing endpoint 0
	vishu.endpoints[1].score = 0

	// Endpoint 0 is on test. Should close the circuit if
	vishu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		assert.Equal(t, "test1", ctx.Target.(string))
		return ctx.Stats, nil
	})

	assert.Equal(t, Closed, int(vishu.endpoints[0].circuitStatus))
}

func TestHalfOpenClosesAgain(t *testing.T) {
	vishu := createInstance(time.Millisecond)

	// Break one endpoint
	vishu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		ctx.Stats["target"] = 1000
		return ctx.Stats, errors.New("Something happened")
	})
	// Wait for the circuit to be half-open
	time.Sleep(time.Millisecond * 3)

	// Check that was switched half-open
	assert.Equal(t, HalfOpen, int(vishu.endpoints[0].circuitStatus))
	assert.Equal(t, Closed, int(vishu.endpoints[1].circuitStatus))

	// Trick to force choosing endpoint 0
	vishu.endpoints[1].score = 0

	// Endpoint 0 is on test. Should re-open on error
	vishu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		assert.Equal(t, "test1", ctx.Target.(string))
		return ctx.Stats, errors.New("something bad happened")
	})

	assert.Equal(t, Open, int(vishu.endpoints[0].circuitStatus))
}
