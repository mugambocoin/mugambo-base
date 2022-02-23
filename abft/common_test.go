package abft

import (
	"github.com/mugambocoin/mugambo-base/inter/idx"
	"github.com/mugambocoin/mugambo-base/inter/pos"
	"github.com/mugambocoin/mugambo-base/kvdb"
	"github.com/mugambocoin/mugambo-base/kvdb/memorydb"
	"github.com/mugambocoin/mugambo-base/mugambobft"
	"github.com/mugambocoin/mugambo-base/utils/adapters"
	"github.com/mugambocoin/mugambo-base/vecfc"
)

type applyBlockFn func(block *mugambobft.Block) *pos.Validators

// TestMugamboBFT extends MugamboBFT for tests.
type TestMugamboBFT struct {
	*IndexedMugamboBFT

	blocks map[idx.Block]*mugambobft.Block

	applyBlock applyBlockFn
}

// FakeMugamboBFT creates empty abft with mem store and equal weights of nodes in genesis.
func FakeMugamboBFT(nodes []idx.ValidatorID, weights []pos.Weight, mods ...memorydb.Mod) (*TestMugamboBFT, *Store, *EventStore) {
	validators := make(pos.ValidatorsBuilder, len(nodes))
	for i, v := range nodes {
		if weights == nil {
			validators[v] = 1
		} else {
			validators[v] = weights[i]
		}
	}

	openEDB := func(epoch idx.Epoch) kvdb.DropableStore {
		return memorydb.New()
	}
	crit := func(err error) {
		panic(err)
	}
	store := NewStore(memorydb.New(), openEDB, crit, LiteStoreConfig())

	err := store.ApplyGenesis(&Genesis{
		Validators: validators.Build(),
		Epoch:      FirstEpoch,
	})
	if err != nil {
		panic(err)
	}

	input := NewEventStore()

	config := LiteConfig()
	lch := NewIndexedMugamboBFT(store, input, &adapters.VectorToDagIndexer{vecfc.NewIndex(crit, vecfc.LiteConfig())}, crit, config)

	extended := &TestMugamboBFT{
		IndexedMugamboBFT: lch,
		blocks:            map[idx.Block]*mugambobft.Block{},
	}

	blockIdx := idx.Block(0)

	err = extended.Bootstrap(mugambobft.ConsensusCallbacks{
		BeginBlock: func(block *mugambobft.Block) mugambobft.BlockCallbacks {
			blockIdx++
			return mugambobft.BlockCallbacks{
				EndBlock: func() (sealEpoch *pos.Validators) {
					// track blocks
					extended.blocks[blockIdx] = block
					if extended.applyBlock != nil {
						return extended.applyBlock(block)
					}
					return nil
				},
			}
		},
	})
	if err != nil {
		panic(err)
	}

	return extended, store, input
}
