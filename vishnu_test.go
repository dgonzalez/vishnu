package vishnu

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createInstance(closingTimeout time.Duration) *Vishnu {
	vishnu := New(func(stats map[string]interface{}) int {
		return stats["target"].(int)
	}, closingTimeout)
	vishnu.Add("test1")
	vishnu.Add("test2")
	return vishnu
}

func TestHappyPath(t *testing.T) {
	vishnu := createInstance(time.Millisecond)
	vishnu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		if ctx.Target.(string) == "test1" {
			ctx.Stats["target"] = 1000
		} else {
			ctx.Stats["target"] = 0
		}
		return ctx.Stats, nil
	})

	vishnu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		assert.Equal(t, ctx.Target.(string), "test1", "best endpoint selected")
		return ctx.Stats, nil
	})
}

func TestNilEndpoint(t *testing.T) {
	vishnu := New(nil, time.Millisecond)
	err := vishnu.Add(nil)
	assert.NotNil(t, err)
}

func TestCircuitOpen(t *testing.T) {
	vishnu := createInstance(time.Millisecond)
	vishnu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		return ctx.Stats, errors.New("Something happened")
	})
	assert.Equal(t, Open, int(vishnu.endpoints[0].circuitStatus))
	assert.Equal(t, Closed, int(vishnu.endpoints[1].circuitStatus))
}

func TestHalfOpenAfterTimeout(t *testing.T) {
	vishnu := createInstance(time.Millisecond)

	// Break one endpoint
	vishnu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		ctx.Stats["target"] = 1000
		return ctx.Stats, errors.New("Something happened")
	})
	// Wait for the circuit to be half-open
	time.Sleep(time.Millisecond * 3)

	// Check that was switched half-open
	assert.Equal(t, HalfOpen, int(vishnu.endpoints[0].circuitStatus))
	assert.Equal(t, Closed, int(vishnu.endpoints[1].circuitStatus))

	// Trick to force choosing enn
	vishnu.endpoints[1].score = 0
	vishnu.With(func(ctx ActionCtx) (map[string]interface{}, error) {
		assert.Equal(t, "test1", ctx.Target.(string))
		return ctx.Stats, nil
	})
	assert.Equal(t, Closed, int(vishnu.endpoints[0].circuitStatus))
}
