package core
import (
	"github.com/DEL-ORG/del/core/state"
	"github.com/DEL-ORG/del/core/types"
	"github.com/DEL-ORG/del/core/vm"
)
type Validator interface {
	ValidateBody(block *types.Block) error
	ValidateState(block, parent *types.Block, state *state.StateDB, receipts types.Receipts, usedGas uint64) error
}
type Processor interface {
	Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, []*types.Log, uint64, error)
}
