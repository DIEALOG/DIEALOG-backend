package crypto

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/DIEALOG/DIEALOG-backend/config"
	"github.com/DIEALOG/DIEALOG-backend/domain"
)

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
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.E.InternalCryptoUrl+"/store", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/xml, application/json, text/xml, */*")
	req.Header.Set("X-Internal-Secret-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf(
			"unexpected status code: %d\nbody: %s",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}

func (c *Client) Retrieve(ctx context.Context, dto domain.GetPasswordRequest) (string, error) {
	reqBody, err := json.Marshal(dto)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.E.InternalCryptoUrl+"/retrieve", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/xml, application/json, text/xml, */*")
	req.Header.Set("X-Internal-Secret-Key", config.E.InternalSecretKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"unexpected status code: %d\nbody: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var res struct {
		AccountName string `json:"account_name"`
		Password    string `json:"password"`
	}
	if err := xml.Unmarshal(body, &res); err != nil {
		return "", err
	}

	if res.AccountName != dto.AccountName {
		return "", fmt.Errorf(
			"invalid account name: %s, expected %s",
			res.AccountName,
			dto.AccountName,
		)
	}

	return res.Password, nil
}
