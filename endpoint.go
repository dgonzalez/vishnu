package vishu

import (
	"errors"
	"time"
)

type circuitStatus int

const (
	Open           = iota
	Closed         = iota
	HalfOpen       = iota
	ClosingTimeout = time.Second * 30
)

type endpoint struct {
	Target        interface{}
	stats         map[string]interface{}
	score         uint
	circuitStatus circuitStatus
}

type Vishu struct {
	endpoints []endpoint
	rater     rater
	action    action
}

type rater func(map[string]interface{}) uint

type action func(interface{}) (map[string]interface{}, error)

func New(rater rater, action action) *Vishu {
	// TODO David: Needs default rater and default action
	vishu := Vishu{nil, rater, action}
	return &vishu
}

func newEndpoint(target interface{}) (*endpoint, error) {
	if target == nil {
		return nil, errors.New("endpoint must not be nil")
	}

	return &endpoint{target, make(map[string]interface{}), 0, Closed}, nil
}

func (v *Vishu) Add(target interface{}) error {
	endpoint, err := newEndpoint(target)
	if err != nil {
		return err
	}
	v.endpoints = append(v.endpoints, *endpoint)

	return nil
}

func (v *Vishu) Choose(action action) {
	var max uint
	var index int
	for i, element := range v.endpoints {
		if element.score > max {
			max = element.score
			index = i
		}
	}

	chosenEndpoint := v.endpoints[index]
	stats, error := v.action(v.endpoints[index])

	if error != nil {
		chosenEndpoint.score = 0
		chosenEndpoint.circuitStatus = Open

		time.AfterFunc(ClosingTimeout, func() {
			chosenEndpoint.circuitStatus = HalfOpen
		})
	}

	chosenEndpoint.score = v.rater(stats)
	chosenEndpoint.stats = stats
}
