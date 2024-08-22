package github

import (
	"bytes"
	"errors"
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

func (c *Client) Post(uri string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create the http request: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gl2gh")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	if c.debug || c.dryRun {
		log.Printf("[http] POST %s\n%s", uri, string(body))
	}

	if c.dryRun {
		return nil, errors.New("No API call will be made (--dry-run switch was specified)")
	}

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

	log.Printf("[http] STATUS %d", resp.StatusCode)

	if resp.StatusCode != http.StatusCreated {
		if c.debug {
			log.Printf("[http] RESPONSE BODY\n%s", respBody)
		}
		return nil, fmt.Errorf("http client: status code: %d", resp.StatusCode)
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
