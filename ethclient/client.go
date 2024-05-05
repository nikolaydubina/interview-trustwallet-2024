package ethclient

import (
	"context"
	"encoding/json"
	"fmt"
)

type JSONRPCClient interface {
	Do(ctx context.Context, method string, params []any) (json.RawMessage, error)
}

type EthereumClient struct {
	client JSONRPCClient
}

func NewEthereumClient(client JSONRPCClient) EthereumClient { return EthereumClient{client: client} }

func (s EthereumClient) GetCurrentBlock(ctx context.Context) (Quantity, error) {
	resp, err := s.client.Do(ctx, "eth_blockNumber", []any{})
	if err != nil {
		return NewQuantityFromInt64(0), fmt.Errorf("cannot make json rpc: %w", err)
	}

	var blockNumber Quantity
	if err := json.Unmarshal(resp, &blockNumber); err != nil {
		return NewQuantityFromInt64(0), err
	}

	return blockNumber, nil
}

// Address supposed to be 20B binary
// TODO: add validation
type Address string

func NewAddressFromString(s string) (Address, error) { return Address(s), nil }

func (s Address) String() string { return string(s) }

// Block details
// TODO: rest of details
type Block struct {
	Number       Quantity      `json:"number"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction details
// TODO: rest of details
type Transaction struct {
	From  Address  `json:"from"` // address of the sender
	To    Address  `json:"to"`   // address of the receiver
	Value Quantity `json:"value"`
}

// GetBlockByNumber with transactions.
func (s EthereumClient) GetBlockByNumber(ctx context.Context, blockNumber Quantity) (*Block, error) {
	resp, err := s.client.Do(ctx, "eth_getBlockByNumber", []any{blockNumber.String(), true})
	if err != nil {
		return nil, fmt.Errorf("cannot make json rpc: %w", err)
	}

	var block Block
	if err := json.Unmarshal(resp, &block); err != nil {
		return nil, err
	}

	return &block, nil
}
