package da

import (
	"context"
	ds "github.com/ipfs/go-datastore"
	"github.com/rollkit/rollkit/log"
	"github.com/rollkit/rollkit/types"
)

type StatusCode uint64

// Data Availability return codes.
const (
	StatusUnknown StatusCode = iota
	StatusSuccess
	StatusNotFound
	StatusError
)

// BaseResult contains basic information returned by DA layer.
type BaseResult struct {
	// Code is to determine if the action succeeded.
	Code StatusCode
	// Message may contain DA layer specific information (like DA block height/hash, detailed error message, etc)
	Message string
	// DAHeight informs about a height on Data Availability Layer for given result.
	DAHeight uint64
}

// ResultSubmitBlock contains information returned from DA layer after block submission.
type ResultSubmitBlock struct {
	BaseResult
}

// ResultCheckBlock contains information about block availability, returned from DA layer client.
type ResultCheckBlock struct {
	BaseResult

	DataAvailable bool
}

// ResultRetrieveBlocks contains batch of blocks returned from DA layer client.
type ResultRetrieveBlocks struct {
	BaseResult

	Blocks []*types.Block
}

// DataAvailabilityLayerClient defines generic interface for DA layer block submission.
// It also contains life-cycle methods.
type DataAvailabilityLayerClient interface {
	// Init is called once to allow DA client to read configuration and initialize resources.
	Init(namespaceID types.NamespaceID, config []byte, kvStore ds.Datastore, logger log.Logger) error

	// Start is called once, after Init. It's implementation should start operation of DataAvailabilityLayerClient.
	Start() error

	// Stop is called once, when DataAvailabilityLayerClient is no longer needed.
	Stop() error

	SubmitBlock(ctx context.Context, block *types.Block) ResultSubmitBlock

	// CheckBlockAvailability queries DA layer to check data availability of block corresponding at given height.
	CheckBlockAvailability(ctx context.Context, dataLayerHeight uint64) ResultCheckBlock
}

// BlockRetriever is additional interface that can be implemented by Data Availability Layer Client that is able to retrieve
// block data from DA layer. This gives the ability to use it for block synchronization.
type BlockRetriever interface {
	// RetrieveBlocks returns blocks at given data layer height from data availability layer.
	RetrieveBlocks(ctx context.Context, dataLayerHeight uint64) ResultRetrieveBlocks
}
