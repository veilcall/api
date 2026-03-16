package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type MoneroClient struct {
	url      string
	user     string
	password string
	httpCli  *http.Client
}

func NewMoneroClient(url, user, password string) *MoneroClient {
	return &MoneroClient{
		url:      url,
		user:     user,
		password: password,
		httpCli:  &http.Client{},
	}
}

type rpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (m *MoneroClient) call(ctx context.Context, method string, params, result interface{}) error {
	body, _ := json.Marshal(rpcRequest{
		JSONRPC: "2.0",
		ID:      "0",
		Method:  method,
		Params:  params,
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, m.url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(m.user, m.password)

	resp, err := m.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("monero rpc: %w", err)
	}
	defer resp.Body.Close()

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("decode monero rpc response: %w", err)
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("monero rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	if result != nil {
		return json.Unmarshal(rpcResp.Result, result)
	}
	return nil
}

type MakeIntegratedAddressResult struct {
	IntegratedAddress string `json:"integrated_address"`
	PaymentID         string `json:"payment_id"`
}

func (m *MoneroClient) MakeIntegratedAddress(ctx context.Context, paymentID string) (*MakeIntegratedAddressResult, error) {
	params := map[string]string{}
	if paymentID != "" {
		params["payment_id"] = paymentID
	}
	var result MakeIntegratedAddressResult
	if err := m.call(ctx, "make_integrated_address", params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

type Transfer struct {
	PaymentID     string  `json:"payment_id"`
	Amount        uint64  `json:"amount"`
	Confirmations uint64  `json:"confirmations"`
	Address       string  `json:"address"`
}

type GetTransfersResult struct {
	In []Transfer `json:"in"`
}

func (m *MoneroClient) GetTransfers(ctx context.Context) ([]Transfer, error) {
	params := map[string]bool{
		"in":      true,
		"pending": false,
		"out":     false,
		"failed":  false,
		"pool":    false,
	}
	var result GetTransfersResult
	if err := m.call(ctx, "get_transfers", params, &result); err != nil {
		return nil, err
	}
	return result.In, nil
}
