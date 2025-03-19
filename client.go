package hago

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Home Assistant API client
type Client struct {
	BaseURL    *url.URL
	APIToken   string
	HTTPClient *http.Client
}

// NewClient creates a new Home Assistant client
func NewClient(baseURL, apiToken string) (*Client, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	return &Client{
		BaseURL:  parsedURL,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: time.Second * 30,
		},
	}, nil
}

// doRequest performs an HTTP request with the proper authentication
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	endpoint, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	requestURL := c.BaseURL.ResolveReference(endpoint)

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, requestURL.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("Content-Type", "application/json")

	return c.HTTPClient.Do(req)
}

// Get sends a GET request to the Home Assistant API
func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post sends a POST request to the Home Assistant API
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.doRequest(http.MethodPost, path, body)
}

// Put sends a PUT request to the Home Assistant API
func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	return c.doRequest(http.MethodPut, path, body)
}

// Delete sends a DELETE request to the Home Assistant API
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}
