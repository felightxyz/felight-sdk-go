package client

import (
	"context"
	"fmt"
	felightApi "github.com/felightxyz/felight-sdk-go/sdk/protobuf/api"
	felightTypes "github.com/felightxyz/felight-sdk-go/sdk/protobuf/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

const (
	SdkVersion = "0.1.0"
	// metadata keys
	sdkVersionHeaderKey = "x-sdk-version"
	authHeaderKey       = "authorization"
)

type TxsRequest struct {
	Filter string
}

type TxsResponse struct {
	Txs []Tx
}

type Tx struct {
	From  []byte
	RawTx []byte
}

type Client struct {
	config *Config
	conn   *grpc.ClientConn
	client felightApi.ApiEthClient
	// metadata to be sent with each request
	md metadata.MD
}

func NewClient(config *Config) (*Client, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	c := &Client{
		config: config,
		md: metadata.New(map[string]string{
			sdkVersionHeaderKey: SdkVersion,
		}),
	}

	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	if conn, err := grpc.NewClient(c.config.Url, c.config.DialOptions()...); err != nil {
		return err
	} else {
		c.conn = conn
		c.client = felightApi.NewApiEthClient(conn)
		return nil
	}
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) SubscribeNewTxs(ctx context.Context, txsRequest *TxsRequest, ch chan<- *TxsResponse) error {
	retry := 0
	ctx = metadata.NewOutgoingContext(ctx, c.md)

	for {
		retry++
		if err := c.innerSubscribeNewTxs(ctx, txsRequest, ch); err != nil {
			if retry > c.config.MaxRetry {
				return fmt.Errorf("SubscribeNewTxs after %v attempts: %w", retry, err)
			}
			time.Sleep(time.Second * time.Duration(c.config.RetryInterval))
		}
	}
}

func (c *Client) innerSubscribeNewTxs(ctx context.Context, txsRequest *TxsRequest, ch chan<- *TxsResponse) error {
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	txsRequestGrpc := &felightTypes.TxsRequest{}
	if txsRequest != nil {
		txsRequestGrpc.Filter = txsRequest.Filter
	}

	res, err := c.client.SubscribeNewTxs(subCtx, txsRequestGrpc)
	if err != nil {
		return err
	}

	for {
		txsResponseGrpc, err := res.Recv()
		if err != nil {
			return err
		}

		txsResponse := TxsResponse{}
		for _, tx := range txsResponseGrpc.Txs {
			tx2 := Tx{}
			tx2.From = tx.From
			tx2.RawTx = tx.RawTx
			txsResponse.Txs = append(txsResponse.Txs, tx2)
		}

		ch <- &txsResponse
	}
}
