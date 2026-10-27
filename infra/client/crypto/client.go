package crypto

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/config"
	"github.com/DIEALOG/DIEALOG-backend/domain"
)

const credentialsPath = "/api/credentials"

type Client struct {
	httpClient *http.Client
	apiKey     string
}

func NewClient(httpClient *http.Client, apiKey string) *Client {
	return &Client{
		httpClient: httpClient,
		apiKey:     apiKey,
	}
}

func (c *Client) Store(ctx context.Context, dto domain.StoreCredentialRequest) error {
	reqBody, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.E.InternalCryptoUrl+credentialsPath+"/store", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Internal-Secret-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf(
			"unexpected status code: %d\nbody: %s",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}

func (c *Client) Retrieve(ctx context.Context, dto domain.GetPasswordRequest) (*string, error) {
	reqBody, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.E.InternalCryptoUrl+credentialsPath+"/retrieve", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Internal-Secret-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		// No credential has been stored for this account yet.
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"unexpected status code: %d\nbody: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var res struct {
		AccountName string  `json:"account_name"`
		Password    *string `json:"password"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	if res.AccountName != dto.AccountName {
		return nil, fmt.Errorf(
			"invalid account name: %s, expected %s",
			res.AccountName,
			dto.AccountName,
		)
	}

	return res.Password, nil
}

func (c *Client) Delete(ctx context.Context, dto domain.DeleteCredentialRequest) error {
	reqBody, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.E.InternalCryptoUrl+credentialsPath+"/delete", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Internal-Secret-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"unexpected status code: %d\nbody: %s",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}
