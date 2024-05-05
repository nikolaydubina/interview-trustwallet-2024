package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	ethclient "github.com/trustwallet/ethparser/ethclient"
	ethparser "github.com/trustwallet/ethparser/ethparser"
	ethparserrepo "github.com/trustwallet/ethparser/ethparser/repository"
	jsonrpc "github.com/trustwallet/ethparser/jsonrpc"
)

type Config struct {
	ServeAddress       string                        `json:"serve_address"`
	EthereumAPIBaseURL string                        `json:"ethereum_api_base_url"`
	RefreshWorker      ethparser.RefreshWorkerConfig `json:"refresh_worker"`
}

func main() {
	configPath := os.Getenv("TRUSTWALLET_FETCHER_CONFIG")
	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("cannot load config from (%s): %s", configPath, err)
	}
	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatalf("cannot parse config: %s", err)
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ethParserRepo := ethparserrepo.NewInMemoryTransactionRepository()

	// worker
	ethClient := ethclient.NewEthereumClient(jsonrpc.NewClient(config.EthereumAPIBaseURL, http.DefaultClient))
	refreshWorker := ethparser.NewRefreshWorker(ethClient, ethParserRepo, config.RefreshWorker)
	go refreshWorker.Run()

	// http server
	ethParser := ethparser.NewParser(ethParserRepo)
	http.HandleFunc("GET /api/v1/current-block", func(w http.ResponseWriter, r *http.Request) {
		block, err := ethParser.GetCurrentBlock(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(block)
	})

	http.HandleFunc("GET /api/v1/{address}/transactions", func(w http.ResponseWriter, r *http.Request) {
		address := r.PathValue("address")
		transactions, err := ethParser.GetTransactions(r.Context(), address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(transactions)
	})

	http.HandleFunc("POST /api/v1/{address}/subscribe", func(w http.ResponseWriter, r *http.Request) {
		address := r.PathValue("address")
		if err := ethParser.Subscribe(r.Context(), address); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	slog.Info("start http server")
	defer slog.Info("stop http server")
	http.ListenAndServe(config.ServeAddress, nil)

	// TODO: finish graceful shutdown
}
