package repository

import (
	"context"
	"log/slog"
	"sync"

	"github.com/trustwallet/ethparser/ethclient"
)

// InMemoryTransactionRepository is basic thread-safe in-memory repository.
// It grows without limit for now.
// Use for testing purposes.
type InMemoryTransactionRepository struct {
	addressSubscription map[string]bool
	transactions        map[string][]ethclient.Transaction
	blockNumber         ethclient.Quantity
	mtx                 *sync.RWMutex // for simplicity locking whole structure in mutex, real DB would not lock all tables on every access!
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		mtx:                 &sync.RWMutex{},
		addressSubscription: make(map[string]bool),
		transactions:        make(map[string][]ethclient.Transaction),
	}
}

func (s *InMemoryTransactionRepository) SetAddressSubscription(ctx context.Context, address ethclient.Address, val bool) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.addressSubscription[address.String()] = val
	return nil
}

func (s *InMemoryTransactionRepository) GetAddressSubscription(ctx context.Context, address ethclient.Address) (bool, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.addressSubscription[address.String()], nil
}

func (s *InMemoryTransactionRepository) GetCurrentBlockNumber(ctx context.Context) (ethclient.Quantity, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.blockNumber, nil
}

func (s *InMemoryTransactionRepository) SetCurrentBlockNumber(ctx context.Context, blockNumber ethclient.Quantity) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.blockNumber = blockNumber
	return nil
}

func (s *InMemoryTransactionRepository) AddTransactionForAddress(ctx context.Context, address ethclient.Address, transaction ethclient.Transaction) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	slog.InfoContext(ctx, "adding transaction", "address", address.String(), "transaction", transaction)
	s.transactions[address.String()] = append(s.transactions[address.String()], transaction)
	return nil
}

func (s *InMemoryTransactionRepository) GetTransactionsForADdress(ctx context.Context, address ethclient.Address) ([]ethclient.Transaction, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	slog.InfoContext(ctx, "get transaction", "address", address.String())
	return s.transactions[address.String()], nil
}
