package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	apiEndpoint    = "https://api.github.com"
	defaultTimeout = 10 * time.Second
)

type RateResponse struct {
	Resources struct {
		Core *Rate `json:"core"`
	} `json:"resources"`
}

type Rate struct {
	Limit     int   `json:"limit"`
	Used      int   `json:"used"`
	Remaining int   `json:"remaining"`
	Reset     int64 `json:"reset"`
	ResetAt   time.Time
}

type Client struct {
	token  string
	client *http.Client
	dryRun bool
	debug  bool
}

func NewClient(token string, dryRun, debug bool) *Client {
	return &Client{
		token:  token,
		debug:  debug,
		dryRun: dryRun,
		client: &http.Client{Timeout: defaultTimeout},
	}
}

func (c *Client) RateLimit() (*Rate, error) {
	req, err := c.NewRequest(http.MethodGet, urljoin(apiEndpoint, "/rate_limit"), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing the request: %v", err)
	}
	resBody, err := c.Do(req, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	rateResponse := &RateResponse{}
	if err := json.Unmarshal(resBody, rateResponse); err != nil {
		return nil, fmt.Errorf("failed to deserialize response: %v", err)
	}
	rate := rateResponse.Resources.Core
	rate.ResetAt = time.Unix(rate.Reset, 0)
	return rate, nil
}

func (c *Client) NewRequest(method, uri string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest(method, uri, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create the http request: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gl2gh")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if c.debug || c.dryRun {
		log.Printf("[http] %s %s\n%s", method, uri, string(body))
	}

	return req, nil
}

func (c *Client) Do(req *http.Request, expectedStatusCode int) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http client: request error: %v", err)
	}

	// resource was created -- read and return the body
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("http client: failed to read response body: %v", err)
	}

	if c.debug {
		log.Printf("[http] STATUS %d", resp.StatusCode)
	}

	if resp.StatusCode != expectedStatusCode {
		if c.debug {
			log.Printf("[http] RESPONSE BODY\n%s", respBody)
		}
		return nil, fmt.Errorf("http client: status code: %d (expected %d)", resp.StatusCode, expectedStatusCode)
	}

	return respBody, nil
}

func urljoin(endpoint, path string) string {
	if strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint[:len(endpoint)-1]
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return endpoint + path
}
