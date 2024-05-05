# interview-trustwallet-2024

by github.com/nikolaydubina

## Goal

Implement Ethereum blockchain parser that will allow to query transactions for subscribed addresses.

## Problem

Users not able to receive push notifications for incoming/outgoing transactions. By Implementing Parser interface we would be able to hook this up to notifications service to notify about any incoming/outgoing transactions.

## Limitations

* Use Go Language
* Avoid usage of external libraries
* Use Ethereum JSONRPC to interact with Ethereum Blockchain
* Use memory storage for storing any data (should be easily extendable to support any storage in the future)

Expose public interface for external usage either via code or command line or rest api that will include supported list of operations defined in the Parser interface

```go
type Parser interface {
    // last parsed block
    GetCurrentBlock() int
    // add address to observer
    Subscribe(address string) bool
    // list of inbound or outbound transactions for an address
    GetTransactions(address string) []Transaction
}
```

## Endpoint

URL: https://cloudflare-eth.com

Request example
```bash
curl -X POST '<https://cloudflare-eth.com>' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}'
```

Result
```json
{
    "id":83,
    "jsonrpc": "2.0",
    "result": "0x4b7" // 1207
}
```

## References

* Ethereum JSON RPC Interface

## Note

* keep it simple
* try to finish the task within 4 hours. We do not track the time spent, this is just a guidance. We do not ask for a perfect production ready service

# Implementation Notes

## Brief context and short architecture:
- there is no API to get transactions or subscribe directly in 3rd party Ethereum API providers
- transactions are bundled into blocks, each 1:1 mapping
- blocks are sequential
- we can get current block number from API
- we can get block transactions from API
- service long-polls data from blocks it did not process so far (it tracks processed blocks so far) and adds transactions to subscribed addresses

## Other notes
- changing signature to return error as well, since we are relying on external systems over network, that we can not guarantee for them to work well neither network. Logic does not seem to be error-proof, meaning errors are possible, thus returning them instead of panic or silent suppression.
- changing signature to add context, that will help for distributed tracing, given we have network calls here
- architecture: we have long-polling process (thread/goroutine) that pools statuses of transactions so far for subscribed addresses and stores them into repository. clients can: 1) start process; 2) add address (to already running process); 3) query results; 4) get current block.
- given long running long-polling server and async interactions with it and its stored data, HTTP serve is best approach, as it elegantly specifies text format of issuing commands to server and getting back results. 
- using vocabulary type for QUANTITY, this is fundamental numeric data type with special encoding across different API fields, thus very useful to keep it in native signatures and API
- there is BlockFilter and Filer API, but it does not seem to fit this problem. More investigation needed. TODO.
- long-polling worker is isolated from user-api class, this can help deployment
- transaction repository is single repo for now, but can be split into multiple ones (Address subscription status, block number, transactions for addresses). for now single class for simplicity, but may as well split. especially given natural different tables for each of data
- standard Go lib implements only JSON RPC v1, so for v2 we need implement ourselves, since can not use 3rd party in solution
- cannot use YAML configs (that have better encoding for duration like `30s`), since requires 3rd party. encoding fields JSON friendly
- cannot use Google API Methods notation, since standard Go mutex does not detect `:` gracefully for custom methods

## Example Run

build and start http server and worker and in-memory db (all one process).
run in root folder, or else provide path to config file via env variable.
```bash
$ TRUSTWALLET_FETCHER_CONFIG=$PWD/config.json go run cmd/server-fetcher/server-fetcher.go 
{"time":"2024-05-05T15:56:07.116309+08:00","level":"INFO","msg":"starting refresh worker"}
{"time":"2024-05-05T15:56:07.116478+08:00","level":"INFO","msg":"start http server"}
{"time":"2024-05-05T15:56:17.518031+08:00","level":"INFO","msg":"previous block is zero, setting current block","block_number":"0x12e29aa"}
{"time":"2024-05-05T15:56:27.450348+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e29aa","to_block":"0x12e29ab"}
{"time":"2024-05-05T15:56:28.071358+08:00","level":"INFO","msg":"processed block","block_number":"0x12e29ab","num_transactions":141}
{"time":"2024-05-05T15:56:28.071418+08:00","level":"INFO","msg":"processed block ok","block_number":"0x12e29ab"}
{"time":"2024-05-05T15:56:37.352099+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e29ab","to_block":"0x12e29ac"}
{"time":"2024-05-05T15:56:38.315916+08:00","level":"INFO","msg":"processed block","block_number":"0x12e29ac","num_transactions":167}
{"time":"2024-05-05T15:56:38.315964+08:00","level":"INFO","msg":"processed block ok","block_number":"0x12e29ac"}
{"time":"2024-05-05T15:56:47.421984+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":0,"from_block":"0x12e29ac","to_block":"0x12e29ac"}
{"time":"2024-05-05T15:56:53.775256+08:00","level":"INFO","msg":"get transaction","address":"0x4675c7e5baafbffbca748158becba61ef3b0a263"}
{"time":"2024-05-05T15:56:57.39537+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e29ac","to_block":"0x12e29ad"}
{"time":"2024-05-05T15:56:58.37303+08:00","level":"INFO","msg":"processed block","block_number":"0x12e29ad","num_transactions":144}
{"time":"2024-05-05T15:56:58.373087+08:00","level":"INFO","msg":"processed block ok","block_number":"0x12e29ad"}
{"time":"2024-05-05T15:57:00.464607+08:00","level":"INFO","msg":"get transaction","address":"0x4675c7e5baafbffbca748158becba61ef3b0a263"}
{"time":"2024-05-05T15:57:07.385928+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e29ad","to_block":"0x12e29ae"}
{"time":"2024-05-05T15:57:08.417306+08:00","level":"INFO","msg":"adding transaction","address":"0x4675c7e5baafbffbca748158becba61ef3b0a263","transaction":{"from":"0x1f9090aae28b8a3dceadf281b0f12828e676c326","to":"0x4675c7e5baafbffbca748158becba61ef3b0a263","value":"0x3960123222d9c5"}}
{"time":"2024-05-05T15:57:08.417434+08:00","level":"INFO","msg":"processed block","block_number":"0x12e29ae","num_transactions":172}
{"time":"2024-05-05T15:57:08.417445+08:00","level":"INFO","msg":"processed block ok","block_number":"0x12e29ae"}
{"time":"2024-05-05T15:57:12.326838+08:00","level":"INFO","msg":"get transaction","address":"0x4675c7e5baafbffbca748158becba61ef3b0a263"}
{"time":"2024-05-05T15:57:17.366714+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e29ae","to_block":"0x12e29af"}
{"time":"2024-05-05T15:57:19.38496+08:00","level":"INFO","msg":"processed block","block_number":"0x12e29af","num_transactions":170}
{"time":"202
```

check current block
```bash
$ curl -X GET 'http://127.0.0.1:8080/api/v1/current-block'
19802355
```

subscribe
```bash
$ curl -X POST 'http://127.0.0.1:8080/api/v1/0x4675c7e5baafbffbca748158becba61ef3b0a263/subscribe'
```

transactions
```bash
$ curl -X GET 'http://127.0.0.1:8080/api/v1/0x4675c7e5baafbffbca748158becba61ef3b0a263/transactions'
[{"from":"0x1f9090aae28b8a3dceadf281b0f12828e676c326","to":"0x4675c7e5baafbffbca748158becba61ef3b0a263","value":"0x3960123222d9c5"}]
```

How to get active addresses that likely to have transactions? (without spending money)

1. get curl latest block `curl -X POST 'https://cloudflare-eth.com' --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":83}'`
2. get some address that has "to" destination in transactions, those addresses likely to make another transaction. `curl -X POST 'https://cloudflare-eth.com' --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x12e29a9", true],"id":1}' | jq`

Other improvements:
* tests. if I had time I would add mock-based tests for classes, for parsing/writing full example responses, but I don't have time! (4h limit and I am moving house! and my laptop is dying)
* OTEL telemetry (extending upon tags I used in slog, but little better)
* CHI HTTP router
* YAML config
* proper database (maybe redis, or maybe async drive architecture to send events from long-polling worker into whoever subscribed through queues AWS SNS+SQS style)
* standard JSON RPC 3rd party lib
* Docker container
* K8S Service and K8S Jobs charts
* separate binary for worker and server
* middleware for HTTP server
* fuzz tests for QUANTITY in ethereum code
* and many more!
