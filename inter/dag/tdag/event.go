package tdag

import (
	"github.com/MugamboBC/mugambo-base/hash"
	"github.com/MugamboBC/mugambo-base/inter/dag"
)

type TestEvent struct {
	dag.MutableBaseEvent
	Name string
}

func (e *TestEvent) AddParent(id hash.Event) {
	parents := e.Parents()
	parents.Add(id)
	e.SetParents(parents)
}
