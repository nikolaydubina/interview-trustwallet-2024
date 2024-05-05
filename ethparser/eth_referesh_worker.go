package ethparser

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/trustwallet/ethparser/ethclient"
)

// RefreshWorker fetches information from Ethereum blockchain and updates repository.
// Safe for concurrent execution, but that may lead to duplicate transactions.
type RefreshWorker struct {
	client ethclient.EthereumClient
	repo   TransactionRepository
	ticker *time.Ticker
}

type RefreshWorkerConfig struct {
	PoolIntervalSec int `json:"pool_interval_sec"`
}

// NewRefreshWorker creates a new worker.
func NewRefreshWorker(client ethclient.EthereumClient, repo TransactionRepository, config RefreshWorkerConfig) RefreshWorker {
	return RefreshWorker{
		client: client,
		repo:   repo,
		ticker: time.NewTicker(time.Second * time.Duration(config.PoolIntervalSec)),
	}
}

func (s RefreshWorker) fetch(ctx context.Context) error {
	prevBlock, err := s.repo.GetCurrentBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("cannot get current block: %w", err)
	}

	blockNumber, err := s.client.GetCurrentBlock(ctx)
	if err != nil {
		return fmt.Errorf("cannot get current block: %w", err)
	}

	if prevBlock.Int64() == 0 {
		slog.InfoContext(ctx, "previous block is zero, setting current block", "block_number", blockNumber.String())
		s.repo.SetCurrentBlockNumber(ctx, blockNumber)
		return nil
	}

	slog.InfoContext(ctx, "processing new blocks", "num_blocks_to_process", blockNumber.Int64()-prevBlock.Int64(), "from_block", prevBlock.String(), "to_block", blockNumber.String())

	for i := prevBlock.Int64() + 1; i <= blockNumber.Int64(); i++ {
		blockNumber := ethclient.NewQuantityFromInt64(i)
		if err := s.processBlock(ctx, blockNumber); err != nil {
			return fmt.Errorf("cannot get block by number: %w", err)
		}
		s.repo.SetCurrentBlockNumber(ctx, blockNumber)
		slog.InfoContext(ctx, "processed block ok", "block_number", blockNumber.String())
	}

	return nil
}

func (s RefreshWorker) processBlock(ctx context.Context, blockNumber ethclient.Quantity) error {
	block, err := s.client.GetBlockByNumber(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("cannot get block by number: %w", err)
	}
	for _, tx := range block.Transactions {
		if err := errors.Join(
			s.repo.AddTransactionForAddress(ctx, tx.From, tx),
			s.repo.AddTransactionForAddress(ctx, tx.To, tx),
		); err != nil {
			return fmt.Errorf("cannot persist transactions: %w", err)
		}
	}
	slog.InfoContext(ctx, "processed block", "block_number", blockNumber.String(), "num_transactions", len(block.Transactions))
	return nil
}

func (s RefreshWorker) Run() {
	slog.Info("starting refresh worker")
	defer slog.Info("stop refresh worker")

	// TODO: graceful shutdown (select based on stop signal)
	// TODO: pass cancelable context
	for range s.ticker.C {
		if err := s.fetch(context.Background()); err != nil {
			slog.Error("cannot fetch data", "error", err)
		}
	}
}
