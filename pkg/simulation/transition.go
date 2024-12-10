package simulation

import (
	"math/rand"
	"time"
)

type Transition interface {
	Next(t time.Time) State
}

type DirectTransition struct {
	next State
}

func Direct(next State) DirectTransition {
	return DirectTransition{
		next: next,
	}
}

func (transition DirectTransition) Next(t time.Time) State {
	return transition.next
}

type ProbabilityTransition struct {
	Splits []Split
}
type Split struct {
	probability float64
	next        Transition
}

func WithProbability(p float64, next Transition) Split {
	return Split{
		probability: p / 100,
		next:        next,
	}
}

func (transition ProbabilityTransition) Next(t time.Time) State {
	for _, split := range transition.Splits {
		if rand.Float64() < split.probability {
			return split.next.Next(t)
		}
	}
	return &END_STATE
}
