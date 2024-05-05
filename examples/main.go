package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log"

	felightSdk "github.com/felightxyz/felight-sdk-go/sdk"
)

var (
	url         = flag.String("url", "", "the url to connect to")
	accessToken = flag.String("accesstoken", "[accessToken]", "the access token to use for authentication")
)

func main() {
	flag.Parse()

	if url == nil || *url == "" {
		panic("url is required")
	}

	config := felightSdk.NewConfig()
	config.Url = *url
	config.AccessToken = *accessToken
	SubscribeTransactions(config, "")
}

func SubscribeTransactions(config *felightSdk.Config, filter string) {
	ch := make(chan *felightSdk.TxsResponse)
	defer close(ch)

	client, err := felightSdk.NewClient(config)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	txsRequest := &felightSdk.TxsRequest{
		Filter: filter,
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := client.SubscribeNewTxs(ctx, txsRequest, ch); err != nil {
			fmt.Printf("Failed to SubscribeNewTxs: %v\n", err)
			panic(err)
		}
	}()

	for txsResponse := range ch {
		for _, tx := range txsResponse.Txs {
			txEth := new(types.Transaction)
			if err := txEth.UnmarshalBinary(tx.RawTx); err != nil {
				continue
			}

			from := common.BytesToAddress(tx.From)
			txHash := txEth.Hash()
			txTo := ""
			if txEth.To() != nil {
				txTo = (txEth.To()).Hex()
			}
			txValue := txEth.Value()
			log.Printf("Hash: %v From: %v To: %v Value: %v",
				txHash, from, txTo, txValue)
		}
	}
}
