package abft

import (
	"github.com/MugamboBC/mugambo-base/hash"
	"github.com/MugamboBC/mugambo-base/inter/dag"
)

// EventSource is a callback for getting events from an external storage.
type EventSource interface {
	HasEvent(hash.Event) bool
	GetEvent(hash.Event) dag.Event
}
