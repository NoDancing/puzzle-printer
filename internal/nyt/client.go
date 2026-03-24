package nyt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	loginURL = "https://myaccount.nytimes.com/svc/ios/v2/login"
)

type Client struct {
	http  *http.Client
	token string
}

func NewClient(email, password, token string) (*Client, error) {
	c := &Client{http: &http.Client{}}
	if token != "" {
		c.token = token
		return c, nil
	}
	if err := c.login(email, password); err != nil {
		return nil, fmt.Errorf("NYT login failed: %w", err)
	}
	return c, nil
}

func (c *Client) login(email, password string) error {
	body := url.Values{
		"login":    {email},
		"password": {password},
	}
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
		Data   struct {
			User struct {
				Token string `json:"token"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding login response: %w", err)
	}
	if result.Data.User.Token == "" {
		return fmt.Errorf("no token in login response (bad credentials?)")
	}
	c.token = result.Data.User.Token
	return nil
}

func (c *Client) get(rawURL string) ([]byte, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cookie", "NYT-S="+c.token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s returned %d", rawURL, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
