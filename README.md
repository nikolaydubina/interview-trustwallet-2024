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
{"time":"2024-05-05T15:28:37.442559+08:00","level":"INFO","msg":"starting refresh worker"}
{"time":"2024-05-05T15:28:37.442664+08:00","level":"INFO","msg":"start http server"}
{"time":"2024-05-05T15:28:47.758005+08:00","level":"INFO","msg":"previous block is zero, setting current block","block_number":"0x12e2920"}
{"time":"2024-05-05T15:28:57.796712+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e2920","to_block":"0x12e2921"}
{"time":"2024-05-05T15:28:59.050138+08:00","level":"INFO","msg":"processed block ok","block_number":"0x12e2921"}
{"time":"2024-05-05T15:29:07.710122+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e2921","to_block":"0x12e2922"}
{"time":"2024-05-05T15:29:08.193103+08:00","level":"INFO","msg":"processed block ok","block_number":"0x12e2922"}
{"time":"2024-05-05T15:29:17.678132+08:00","level":"INFO","msg":"processing new blocks","num_blocks_to_process":1,"from_block":"0x12e2922","to_block":"0x12e2923"}
{"time":"2024-05-05T15:29:18.313776+08:00","level":"INFO","msg":"processed block ok","block_number":"0x12e2923"}
...
```

check current block
```bash
$ curl -X GET 'http://127.0.0.1:8080/api/v1/current-block'
19802355
```

subscribe
```bash
$ curl -X POST 'http://127.0.0.1:8080/api/v1/0x82c917933a7b730ce50a13f753aef81a8ff9d7a8/subscribe'
```

transactions
```bash
$ curl -X POST 'http://127.0.0.1:8080/api/v1/0x82c917933a7b730ce50a13f753aef81a8ff9d7a8/transactions'
```
