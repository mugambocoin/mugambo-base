package abft

import (
	"github.com/MugamboBC/mugambo-base/abft/dagidx"
	"github.com/MugamboBC/mugambo-base/hash"
	"github.com/MugamboBC/mugambo-base/inter/dag"
	"github.com/MugamboBC/mugambo-base/inter/idx"
	"github.com/MugamboBC/mugambo-base/inter/pos"
	"github.com/MugamboBC/mugambo-base/mugambobft"
)

var _ mugambobft.Consensus = (*MugamboBFT)(nil)

type DagIndex interface {
	dagidx.VectorClock
	dagidx.ForklessCause
}

// MugamboBFT performs events ordering and detects cheaters
// It's a wrapper around Orderer, which adds features which might potentially be application-specific:
// confirmed events traversal, cheaters detection.
// Use this structure if need a general-purpose consensus. Instead, use lower-level abft.Orderer.
type MugamboBFT struct {
	*Orderer
	dagIndex      DagIndex
	uniqueDirtyID uniqueID
	callback      mugambobft.ConsensusCallbacks
}

// New creates MugamboBFT instance.
func NewMugamboBFT(store *Store, input EventSource, dagIndex DagIndex, crit func(error), config Config) *MugamboBFT {
	p := &MugamboBFT{
		Orderer:  NewOrderer(store, input, dagIndex, crit, config),
		dagIndex: dagIndex,
	}

	return p
}

func (p *MugamboBFT) confirmEvents(frame idx.Frame, atropos hash.Event, onEventConfirmed func(dag.Event)) error {
	err := p.dfsSubgraph(atropos, func(e dag.Event) bool {
		decidedFrame := p.store.GetEventConfirmedOn(e.ID())
		if decidedFrame != 0 {
			return false
		}
		// mark all the walked events as confirmed
		p.store.SetEventConfirmedOn(e.ID(), frame)
		if onEventConfirmed != nil {
			onEventConfirmed(e)
		}
		return true
	})
	return err
}

func (p *MugamboBFT) applyAtropos(decidedFrame idx.Frame, atropos hash.Event) *pos.Validators {
	atroposVecClock := p.dagIndex.GetMergedHighestBefore(atropos)

	validators := p.store.GetValidators()
	// cheaters are ordered deterministically
	cheaters := make([]idx.ValidatorID, 0, validators.Len())
	for creatorIdx, creator := range validators.SortedIDs() {
		if atroposVecClock.Get(idx.Validator(creatorIdx)).IsForkDetected() {
			cheaters = append(cheaters, creator)
		}
	}

	if p.callback.BeginBlock == nil {
		return nil
	}
	blockCallback := p.callback.BeginBlock(&mugambobft.Block{
		Atropos:  atropos,
		Cheaters: cheaters,
	})

	// traverse newly confirmed events
	err := p.confirmEvents(decidedFrame, atropos, blockCallback.ApplyEvent)
	if err != nil {
		p.crit(err)
	}

	if blockCallback.EndBlock != nil {
		return blockCallback.EndBlock()
	}
	return nil
}

func (p *MugamboBFT) Bootstrap(callback mugambobft.ConsensusCallbacks) error {
	return p.BootstrapWithOrderer(callback, p.OrdererCallbacks())
}

func (p *MugamboBFT) BootstrapWithOrderer(callback mugambobft.ConsensusCallbacks, ordererCallbacks OrdererCallbacks) error {
	err := p.Orderer.Bootstrap(ordererCallbacks)
	if err != nil {
		return err
	}
	p.callback = callback
	return nil
}

func (p *MugamboBFT) OrdererCallbacks() OrdererCallbacks {
	return OrdererCallbacks{
		ApplyAtropos: p.applyAtropos,
	}
}
