package lightning

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"PamojaBuild/internal/config"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) (*Client, error) {
	cert, err := os.ReadFile(cfg.LNDTLSCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(cert) {
		return nil, fmt.Errorf("failed to parse TLS cert")
	}

	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: certPool,
				},
			},
			Timeout: 30 * time.Second,
		},
	}, nil
}

type invoiceResponse struct {
	PaymentRequest string `json:"payment_request"`
	PaymentHash    string `json:"r_hash"`
}

type payInvoiceResponse struct {
	TxID string `json:"txid"`
}

type paymentStatusResponse struct {
	Settled bool `json:"settled"`
}

func (c *Client) CreateInvoice(amountSats int64, memo string) (paymentRequest string, paymentHash string, err error) {
	url := fmt.Sprintf("https://%s/v1/invoices", c.config.LNDHost)

	reqBody := map[string]interface{}{
		"value":  amountSats,
		"memo":   memo,
		"expiry": 3600,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/json")
	c.setMacaroonHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to create invoice: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("create invoice failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result invoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("failed to decode invoice response: %w", err)
	}

	// LND returns r_hash as base64-encoded, convert to hex for storage
	hashBytes, err := base64.StdEncoding.DecodeString(result.PaymentHash)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode payment hash from base64: %w", err)
	}

	paymentHash = hex.EncodeToString(hashBytes)
	paymentRequest = result.PaymentRequest

	return paymentRequest, paymentHash, nil
}

func (c *Client) PayInvoice(paymentRequest string) error {
	url := fmt.Sprintf("https://%s/v1/pay", c.config.LNDHost)

	reqBody := map[string]interface{}{
		"payment_request": paymentRequest,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	c.setMacaroonHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to pay invoice: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("pay invoice failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) CheckPaymentStatus(paymentHash string) (settled bool, err error) {
	// paymentHash comes in as hex, LND needs base64
	hashBytes, err := hex.DecodeString(paymentHash)
	if err != nil {
		return false, fmt.Errorf("failed to decode payment hash from hex: %w", err)
	}
	hashBase64 := base64.StdEncoding.EncodeToString(hashBytes)

	url := fmt.Sprintf("https://%s/v1/invoices/%s", c.config.LNDHost, hashBase64)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	c.setMacaroonHeader(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check payment status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("check payment status failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result paymentStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode payment status response: %w", err)
	}

	return result.Settled, nil
}

func (c *Client) setMacaroonHeader(req *http.Request) {
	req.Header.Set("Grpc-Metadata-Macaroon", c.config.LNDMacaroonHex)
}