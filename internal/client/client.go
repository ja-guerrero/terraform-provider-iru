package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Client holds the configuration for the Iru API client.
type Client struct {
	HTTPClient *http.Client
	APIURL     string
	APIToken   string
}

// NewClient creates a new Iru API client.
func NewClient(apiURL, apiToken string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		APIURL:     strings.TrimSuffix(apiURL, "/"),
		APIToken:   apiToken,
	}
}

// DoRequest performs an HTTP request to the Iru API.
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}, response interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.APIURL, path), reqBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	if response != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	}

	return nil
}

// DoMultipartRequest performs a multipart/form-data request to the Iru API.
func (c *Client) DoMultipartRequest(ctx context.Context, method, path string, fields map[string]string, fileField, fileName string, fileContent io.Reader, response interface{}) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, val := range fields {
		_ = writer.WriteField(key, val)
	}

	if fileContent != nil {
		part, err := writer.CreateFormFile(fileField, fileName)
		if err != nil {
			return fmt.Errorf("error creating form file: %w", err)
		}
		_, err = io.Copy(part, fileContent)
		if err != nil {
			return fmt.Errorf("error copying file content: %w", err)
		}
	}

	err := writer.Close()
	if err != nil {
		return fmt.Errorf("error closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.APIURL, path), body)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	if response != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
	}

	return nil
}
