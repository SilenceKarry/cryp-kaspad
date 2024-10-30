package client

import (
	"bytes"
	"context"
	"cryp-kaspad/internal/libs/eos/core/types"
	"encoding/json"
	"fmt"

	"io"
	"net/http"
	"net/url"
	"strings"

	eos "github.com/eoscanada/eos-go"
)

type Client struct {
	c    *http.Client
	API  *eos.API
	host string
}

type CreatorConfig struct {
	PrivateKey string
	AccountID  string
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*Client, error) {
	return DialContext(context.Background(), rawurl)
}

// DialContext connects a client to the given URL with context.
func DialContext(ctx context.Context, host string) (*Client, error) {
	c := &http.Client{}
	h, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(h.Path, "/") {
		h.Path += "/"
	}

	return &Client{
		c:    c,
		host: h.String(),
		API:  eos.New(h.String()),
	}, nil
}

// Client gets the underlying RPC client.
func (ec *Client) Client() *http.Client {
	return ec.c
}

func (ec *Client) do(ctx context.Context, path string, method string, result interface{}, data interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, ec.host+path, nil)
	if err != nil {
		return err
	}

	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
		if data != nil {
			body, err := json.Marshal(data)
			if err != nil {
				return err
			}

			req.Body = io.NopCloser(bytes.NewReader(body))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}
	}

	resp, err := ec.c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("http error: %s\n%s", resp.Status, body)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }

	// fmt.Printf("body: %s\n", body)

	// err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	return nil
}

func (ec *Client) GetInfo(ctx context.Context) (*types.ChainInfo, error) {
	var info types.ChainInfo
	err := ec.do(ctx, "v1/chain/get_info", "GET", &info, nil)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (ec *Client) GetBlock(ctx context.Context, blockNum uint32) (*types.Block, error) {
	var block types.Block
	err := ec.do(ctx, "v1/chain/get_block", "POST", &block, map[string]interface{}{
		"block_num_or_id": blockNum,
	})
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (ec *Client) GetAccount(ctx context.Context, accountName string) (*types.Account, error) {
	var account types.Account
	err := ec.do(ctx, "v1/chain/get_account", "POST", &account, map[string]interface{}{
		"account_name": accountName,
	})
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (ec *Client) GetTokenBalance(ctx context.Context, accountName string, contractName string, symbol string) ([]string, error) {
	assets, err := ec.API.GetCurrencyBalance(ctx, eos.AN(accountName), symbol, eos.AN(contractName))
	if err != nil {
		return []string{}, err
	}

	results := make([]string, len(assets))
	for i, asset := range assets {
		results[i] = asset.String()
	}

	return results, nil
}

func (ec *Client) ImportPrivateKey(ctx context.Context, privateKey string) error {
	if privateKey == "" {
		return fmt.Errorf("private key is empty")
	}

	keyBag := &eos.KeyBag{}
	err := keyBag.ImportPrivateKey(ctx, privateKey)
	if err != nil {
		return err
	}

	ec.API.SetSigner(keyBag)
	return nil
}
