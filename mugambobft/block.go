package mugambobft

import (
	"github.com/mugambocoin/mugambo-base/hash"
)

// Block is a part of an ordered chain of batches of events.
type Block struct {
	Atropos  hash.Event
	Cheaters Cheaters
}
