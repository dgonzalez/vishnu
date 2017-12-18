package main

import (
	"errors"
	"math"
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
	// ClosingTimeout is the default HalfOpen to Closed timeout.
	ClosingTimeout = time.Second * 30
)

type endpoint struct {
	Target        interface{}
	stats         map[string]interface{}
	score         int
	circuitStatus circuitStatus
}

// Vishu instance. It is the core of the module.
type Vishu struct {
	endpoints []endpoint
	rater     rater
}

type rater func(map[string]interface{}) int

func defaultRater(stats map[string]interface{}) int {
	execTime := stats["__execution_time"]
	if execTime != nil {
		return math.MaxInt32 - int(execTime.(time.Duration))
	}
	return 0
}

type action func(interface{}) (map[string]interface{}, error)

// New creates a Vishu instance with a custom rating function.
func New(rater rater) *Vishu {
	// TODO David: Needs default rater and default action
	vishu := Vishu{nil, rater}
	return &vishu
}

func newEndpoint(target interface{}) (*endpoint, error) {
	if target == nil {
		return nil, errors.New("endpoint must not be nil")
	}

	return &endpoint{target, make(map[string]interface{}), 500, Closed}, nil
}

// Add adds an endpoint
func (v *Vishu) Add(target interface{}) error {
	endpoint, err := newEndpoint(target)
	if err != nil {
		return err
	}
	v.endpoints = append(v.endpoints, *endpoint)
	return nil
}

// With selects an endpoint
func (v *Vishu) With(action action) {
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
	stats, error := action(v.endpoints[index].Target)
	finish := time.Now()
	elapsed := finish.Sub(start)
	stats["__execution_time"] = elapsed

	if error != nil {
		chosenEndpoint.score = 0
		chosenEndpoint.circuitStatus = Open

		time.AfterFunc(ClosingTimeout, func() {
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
