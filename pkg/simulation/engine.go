package simulation

import "time"

type Engine struct {
	t       time.Time
	current []State
	// record  []*fhir.Resource
	// state map[string]fhir.Value
}

func New() Engine {
	return Engine{}
}

func (eng *Engine) Advance(step time.Duration) {
	eng.t = eng.t.Add(step)
	for i, cur := range eng.current {
		eng.current[i] = cur.Next(eng.t)
	}
}
