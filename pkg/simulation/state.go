package simulation

import (
	"fmt"
	"time"
)

type StateBase struct {
	name       string
	transition Transition
}

func (base *StateBase) Next(t time.Time) State {
	return base.transition.Next(t)
}

type State interface {
	fmt.Stringer
	Next(t time.Time) State
}

type End struct{ StateBase }

func EndState() End {
	end := End{}
	end.transition = Direct(&end)
	return end
}

var END_STATE = EndState()

func (End) String() string {
	return "< End >"
}

type SimpleState struct{ StateBase }

func (state SimpleState) String() string {
	return state.name
}

func Simple(name string, transition Transition) SimpleState {
	return SimpleState{StateBase: StateBase{
		name:       name,
		transition: transition,
	}}
}

type DelayState struct {
	delay  time.Duration
	target time.Time
	StateBase
}

func (state DelayState) String() string {
	return fmt.Sprintf("%s [Delay %s]", state.name, state.delay)
}

func (state DelayState) Next(t time.Time) State {
	if state.target.IsZero() {
		state.target = t
	}
	if t.Compare(state.target) < 0 {
		return state
	}
	return state.StateBase.Next(t)
}
