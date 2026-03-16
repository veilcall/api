package number

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const telnyxBaseURL = "https://api.telnyx.com/v2"

type TelnyxClient struct {
	apiKey  string
	httpCli *http.Client
}

func NewTelnyxClient(apiKey string) *TelnyxClient {
	return &TelnyxClient{apiKey: apiKey, httpCli: &http.Client{}}
}

func (t *TelnyxClient) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, telnyxBaseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+t.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpCli.Do(req)
	if err != nil {
		return fmt.Errorf("telnyx request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telnyx API error: %d", resp.StatusCode)
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

type SearchNumbersResult struct {
	Data []struct {
		PhoneNumber string `json:"phone_number"`
	} `json:"data"`
}

func (t *TelnyxClient) SearchNumbers(ctx context.Context, countryCode string) ([]string, error) {
	path := fmt.Sprintf("/available_phone_numbers?filter[country_code]=%s&filter[limit]=5", countryCode)
	var result SearchNumbersResult
	if err := t.doRequest(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	numbers := make([]string, 0, len(result.Data))
	for _, d := range result.Data {
		numbers = append(numbers, d.PhoneNumber)
	}
	return numbers, nil
}

type OrderNumberResult struct {
	Data struct {
		PhoneNumbers []struct {
			PhoneNumber string `json:"phone_number"`
		} `json:"phone_numbers"`
	} `json:"data"`
}

func (t *TelnyxClient) OrderNumber(ctx context.Context, phoneNumber string) error {
	body := map[string]interface{}{
		"phone_numbers": []map[string]string{
			{"phone_number": phoneNumber},
		},
	}
	return t.doRequest(ctx, http.MethodPost, "/number_orders", body, nil)
}

func (t *TelnyxClient) ReleaseNumber(ctx context.Context, phoneNumber string) error {
	path := fmt.Sprintf("/phone_numbers/%s", phoneNumber)
	return t.doRequest(ctx, http.MethodDelete, path, nil, nil)
}
