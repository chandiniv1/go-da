package avail

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/chandiniv1/go-da/da"
	"github.com/chandiniv1/go-da/da/datasubmit"
	ds "github.com/ipfs/go-datastore"
	"github.com/rollkit/rollkit/log"
	"github.com/rollkit/rollkit/types"
)

type Config struct {
	BaseURL    string  `json:"base_url"`
	Seed       string  `json:"seed"`
	ApiURL     string  `json:"api_url"`
	AppID      int     `json:"app_id"`
	Confidence float64 `json:"confidence"`
}

type DataAvailabilityLayerClient struct {
	namespace types.NamespaceID
	config    Config
	logger    log.Logger
}

type Confidence struct {
	Block                uint32  `json:"block"`
	Confidence           float64 `json:"confidence"`
	SerialisedConfidence *string `json:"serialised_confidence,omitempty"`
}

type AppData struct {
	Block      uint32   `json:"block"`
	Extrinsics []string `json:"extrinsics"`
}

type InitRequest struct {
	namespaceID types.NamespaceID
	config      []byte
	kvStore     ds.Datastore
	logger      log.Logger
}

type InitResponse struct {
	err error
}

// Init initializes DataAvailabilityLayerClient instance.
func (c *DataAvailabilityLayerClient) Init(request InitRequest, reply InitResponse) error {
	c.logger = request.logger

	if len(request.config) > 0 {
		return json.Unmarshal(request.config, &c.config)
	}

	return nil
}

// Start prepares DataAvailabilityLayerClient to work.
func (c *DataAvailabilityLayerClient) Start() error {

	c.logger.Info("starting avail Data Availability Layer Client", "baseURL", c.config.ApiURL)

	return nil
}

// Stop stops DataAvailabilityLayerClient.
func (c *DataAvailabilityLayerClient) Stop() error {

	c.logger.Info("stopping Avail Data Availability Layer Client")

	return nil
}

// SubmitBlock submits a block to DA layer.
func (c *DataAvailabilityLayerClient) SubmitBlock(ctx context.Context, block *types.Block) da.ResultSubmitBlock {

	data, err := block.MarshalBinary()
	if err != nil {
		return da.ResultSubmitBlock{
			BaseResult: da.BaseResult{
				Code:    da.StatusError,
				Message: err.Error(),
			},
		}
	}

	txHash, err := datasubmit.SubmitData(c.config.ApiURL, c.config.Seed, c.config.AppID, data)

	if err != nil {
		return da.ResultSubmitBlock{
			BaseResult: da.BaseResult{
				Code:    da.StatusError,
				Message: err.Error(),
			},
		}
	}

	return da.ResultSubmitBlock{
		BaseResult: da.BaseResult{
			Code:     da.StatusSuccess,
			Message:  "tx hash: " + hex.EncodeToString(txHash[:]),
			DAHeight: 1,
		},
	}
}

// CheckBlockAvailability queries DA layer to check data availability of block.
func (c *DataAvailabilityLayerClient) CheckBlockAvailability(ctx context.Context, dataLayerHeight uint64) da.ResultCheckBlock {

	blockNumber := dataLayerHeight
	confidenceURL := fmt.Sprintf(c.config.BaseURL+"/confidence/%d", blockNumber)

	response, err := http.Get(confidenceURL)

	if err != nil {
		return da.ResultCheckBlock{
			BaseResult: da.BaseResult{
				Code:    da.StatusError,
				Message: err.Error(),
			},
		}
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return da.ResultCheckBlock{
			BaseResult: da.BaseResult{
				Code:    da.StatusError,
				Message: err.Error(),
			},
		}
	}

	var confidenceObject Confidence
	err = json.Unmarshal(responseData, &confidenceObject)
	if err != nil {
		return da.ResultCheckBlock{
			BaseResult: da.BaseResult{
				Code:    da.StatusError,
				Message: err.Error(),
			},
		}
	}

	return da.ResultCheckBlock{
		BaseResult: da.BaseResult{
			Code:     da.StatusSuccess,
			DAHeight: uint64(confidenceObject.Block),
		},
		DataAvailable: confidenceObject.Confidence > float64(c.config.Confidence),
	}
}

//RetrieveBlocks gets the block from DA layer.

func (c *DataAvailabilityLayerClient) RetrieveBlocks(ctx context.Context, dataLayerHeight uint64) da.ResultRetrieveBlocks {
	blocks := []*types.Block{}

Loop:
	blockNumber := dataLayerHeight
	appDataURL := fmt.Sprintf(c.config.BaseURL+"/appdata/%d?decode=true", blockNumber)
	response, err := http.Get(appDataURL)
	if err != nil {
		return da.ResultRetrieveBlocks{
			BaseResult: da.BaseResult{
				Code:    da.StatusError,
				Message: err.Error(),
			},
		}
	}
	responseData, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return da.ResultRetrieveBlocks{
			BaseResult: da.BaseResult{
				Code:    da.StatusError,
				Message: err.Error(),
			},
		}
	}

	var appDataObject AppData

	if string(responseData) == "\"Not found\"" {
		appDataObject = AppData{Block: uint32(blockNumber), Extrinsics: []string{}}
	} else if string(responseData) == "\"Processing block\"" {
		goto Loop
	} else {
		err := json.Unmarshal(responseData, &appDataObject)
		if err != nil {
			fmt.Println(string(responseData))
			return da.ResultRetrieveBlocks{
				BaseResult: da.BaseResult{
					Code:    da.StatusError,
					Message: err.Error(),
				},
			}
		}
	}

	txnsByteArray := []byte{}
	for _, extrinsic := range appDataObject.Extrinsics {
		txnsByteArray = append(txnsByteArray, []byte(extrinsic)...)
	}

	block := &types.Block{
		SignedHeader: types.SignedHeader{
			Header: types.Header{
				BaseHeader: types.BaseHeader{
					Height: blockNumber,
				},
			}},
		Data: types.Data{
			Txs: types.Txs{txnsByteArray},
		},
	}
	blocks = append(blocks, block)

	return da.ResultRetrieveBlocks{
		BaseResult: da.BaseResult{
			Code:     da.StatusSuccess,
			DAHeight: uint64(appDataObject.Block),
		},
		Blocks: blocks,
	}
}
