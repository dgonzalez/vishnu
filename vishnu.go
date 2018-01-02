package vishnu

import (
	"errors"
	"time"
)

type circuitStatus int

const (
	// Open circuit status.
	Open = iota
	// Closed circuit status.
	Closed = iota
	// HalfOpen circuitStatus
	HalfOpen = iota
)

// Vishnu instance. It is the core of the module.
type Vishnu struct {
	endpoints      []endpoint
	rater          rater
	closingTimeout time.Duration
}

// ActionCtx context to simplify calls to actions.
type ActionCtx struct {
	Target interface{}
	Stats  map[string]interface{}
}

type endpoint struct {
	Target        interface{}
	stats         map[string]interface{}
	score         int
	circuitStatus circuitStatus
}

type rater func(map[string]interface{}) int

type action func(ctx ActionCtx) (map[string]interface{}, error)

// New creates a Vishnu instance with a custom rating function.
func New(rater rater, closingTimeout time.Duration) *Vishnu {
	// TODO David: Needs default rater based on __execution_time
	vishnu := Vishnu{nil, rater, closingTimeout}
	return &vishnu
}

func newEndpoint(target interface{}) (*endpoint, error) {
	if target == nil {
		return nil, errors.New("endpoint must not be nil")
	}

	return &endpoint{target, make(map[string]interface{}), 500, Closed}, nil
}

// Add adds an endpoint
func (v *Vishnu) Add(target interface{}) error {
	endpoint, err := newEndpoint(target)
	if err != nil {
		return err
	}
	v.endpoints = append(v.endpoints, *endpoint)
	return nil
}

// With selects an endpoint and executes the action
func (v *Vishnu) With(action action) {
	var max, index int
	for i, element := range v.endpoints {
		if element.circuitStatus != Open && element.score > max {
			max = element.score
			index = i
		}
	}

	chosenEndpoint := &v.endpoints[index]

	// Measure the execution time as a default metric
	start := time.Now()
	stats, error := action(ActionCtx{chosenEndpoint.Target, chosenEndpoint.stats})
	finish := time.Now()
	elapsed := finish.Sub(start)
	stats["__execution_time"] = elapsed

	if error != nil {
		chosenEndpoint.score = 0
		chosenEndpoint.circuitStatus = Open

		time.AfterFunc(v.closingTimeout, func() {
			chosenEndpoint.circuitStatus = HalfOpen
		})
	} else {
		if chosenEndpoint.circuitStatus == HalfOpen {
			chosenEndpoint.score = 500
			chosenEndpoint.circuitStatus = Closed
		}
		// If there is an error, no rating happens.
		chosenEndpoint.score = v.rater(stats)
		chosenEndpoint.stats = stats
	}
}

func main() {

}
