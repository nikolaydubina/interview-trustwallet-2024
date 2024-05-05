package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client implements JSON RPC communication over HTTP that follows standard encoding scheme and method invocation.
type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string, client *http.Client) Client {
	return Client{baseURL: baseURL, client: client}
}

// Do encodes data into JSON RPC payload, calls specified method and returns raw un-decoded data.
// Caller is responsible for decoding response data into correct types.
// Caller is responsible for encoding request parameters.
// This class is expected to handle JSON RPC "id" handling.
func (s Client) Do(ctx context.Context, method string, params []any) (json.RawMessage, error) {
	// TODO: use random ID
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      10,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Post(s.baseURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("cannot make request to (%s) with error: %w", s.baseURL, err)
	}
	defer resp.Body.Close()

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("cannot decode response with error: %w", err)
	}

	return respBody.Result, nil
}

// Request is generic container for JSON RPC protocol requests.
type Request struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

// Response is generic container for JSON RPC protocol responses.
type Response struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
}
