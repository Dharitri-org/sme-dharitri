package blockAPI

import (
	"encoding/hex"

	apiBlock "github.com/Dharitri-org/sme-dharitri/api/block"
	"github.com/Dharitri-org/sme-dharitri/core"
	"github.com/Dharitri-org/sme-dharitri/data/block"
	"github.com/Dharitri-org/sme-dharitri/dataRetriever"
)

type metaAPIBlockProcessor struct {
	*baseAPIBockProcessor
}

// NewMetaApiBlockProcessor will create a new instance of meta api block processor
func NewMetaApiBlockProcessor(arg *APIBlockProcessorArg) *metaAPIBlockProcessor {
	isFullHistoryNode := arg.HistoryRepo.IsEnabled()
	return &metaAPIBlockProcessor{
		baseAPIBockProcessor: &baseAPIBockProcessor{
			isFullHistoryNode:        isFullHistoryNode,
			selfShardID:              arg.SelfShardID,
			store:                    arg.Store,
			marshalizer:              arg.Marshalizer,
			uint64ByteSliceConverter: arg.Uint64ByteSliceConverter,
			historyRepo:              arg.HistoryRepo,
			unmarshalTx:              arg.UnmarshalTx,
		},
	}
}

// GetBlockByNonce wil return a meta APIBlock by nonce
func (mbp *metaAPIBlockProcessor) GetBlockByNonce(nonce uint64, withTxs bool) (*apiBlock.APIBlock, error) {
	storerUnit := dataRetriever.MetaHdrNonceHashDataUnit

	nonceToByteSlice := mbp.uint64ByteSliceConverter.ToByteSlice(nonce)
	headerHash, err := mbp.store.Get(storerUnit, nonceToByteSlice)
	if err != nil {
		return nil, err
	}

	blockBytes, err := mbp.getFromStorer(dataRetriever.MetaBlockUnit, headerHash)
	if err != nil {
		return nil, err
	}

	return mbp.convertMetaBlockBytesToAPIBlock(headerHash, blockBytes, withTxs)
}

// GetBlockByHash will return a shard APIBlock by hash
func (mbp *metaAPIBlockProcessor) GetBlockByHash(hash []byte, withTxs bool) (*apiBlock.APIBlock, error) {
	blockBytes, err := mbp.getFromStorer(dataRetriever.BlockHeaderUnit, hash)
	if err != nil {
		return nil, err
	}

	return mbp.convertMetaBlockBytesToAPIBlock(hash, blockBytes, withTxs)
}

func (mbp *metaAPIBlockProcessor) convertMetaBlockBytesToAPIBlock(hash []byte, blockBytes []byte, withTxs bool) (*apiBlock.APIBlock, error) {
	blockHeader := &block.MetaBlock{}
	err := mbp.marshalizer.Unmarshal(blockHeader, blockBytes)
	if err != nil {
		return nil, err
	}

	headerEpoch := blockHeader.Epoch

	numOfTxs := uint32(0)
	miniblocks := make([]*apiBlock.APIMiniBlock, 0)
	for _, mb := range blockHeader.MiniBlockHeaders {
		if mb.Type == block.PeerBlock {
			continue
		}

		numOfTxs += mb.TxCount

		miniblockAPI := &apiBlock.APIMiniBlock{
			Hash:               hex.EncodeToString(mb.Hash),
			Type:               mb.Type.String(),
			SourceShardID:      mb.SenderShardID,
			DestinationShardID: mb.ReceiverShardID,
		}
		if withTxs {
			miniblockAPI.Transactions = mbp.getTxsByMb(&mb, headerEpoch)
		}

		miniblocks = append(miniblocks, miniblockAPI)
	}

	shardHeaderHashes := make([]string, len(blockHeader.ShardInfo))
	for idx := 0; idx < len(blockHeader.ShardInfo); idx++ {
		shardHeaderHashes[idx] = hex.EncodeToString(blockHeader.ShardInfo[idx].HeaderHash)
	}

	return &apiBlock.APIBlock{
		Nonce:                blockHeader.Nonce,
		Round:                blockHeader.Round,
		Epoch:                blockHeader.Epoch,
		ShardID:              core.MetachainShardId,
		Hash:                 hex.EncodeToString(hash),
		NumTxs:               numOfTxs,
		NotarizedBlockHashes: shardHeaderHashes,
		MiniBlocks:           miniblocks,
	}, nil
}
