package ethparser

import (
	"context"
	"fmt"

	"github.com/trustwallet/ethparser/ethclient"
)

type TransactionRepository interface {
	EnableAddressSubscription(ctx context.Context, address ethclient.Address) error
	GetSubscribedAddresses(ctx context.Context) ([]ethclient.Address, error)
	GetCurrentBlockNumber(ctx context.Context) (ethclient.Quantity, error)
	SetCurrentBlockNumber(ctx context.Context, blockNumber ethclient.Quantity) error
	AddTransactionForAddress(ctx context.Context, address ethclient.Address, transaction ethclient.Transaction) error
	GetTransactionsForADdress(ctx context.Context, address ethclient.Address) ([]ethclient.Transaction, error)
}

// Parser tracks subscribed addresses from Ethereum blockchain by long-polling.
// This is client api (read path).
// To insert data, refresh worker has to run and connected to same repository (write path).
type Parser struct {
	repository TransactionRepository
}

func NewParser(repository TransactionRepository) Parser {
	return Parser{repository: repository}
}

// GetCurrentBlock which is a the last parsed block.
func (s Parser) GetCurrentBlock(ctx context.Context) (int, error) {
	v, err := s.repository.GetCurrentBlockNumber(ctx)
	return int(v.Int64()), err
}

// Subscribe adds address to observer.
func (s Parser) Subscribe(ctx context.Context, address string) error {
	addr, err := ethclient.NewAddressFromString(address)
	if err != nil {
		return fmt.Errorf("invalid address: %w", err)
	}
	return s.repository.EnableAddressSubscription(ctx, addr)
}

// GetTransactions lists of inbound or outbound transactions for an address.
func (s Parser) GetTransactions(ctx context.Context, address string) ([]ethclient.Transaction, error) {
	addr, err := ethclient.NewAddressFromString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}
	return s.repository.GetTransactionsForADdress(ctx, addr)
}
