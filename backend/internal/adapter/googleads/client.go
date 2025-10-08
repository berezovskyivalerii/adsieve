package googleads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type linkRepo interface {
    LinkGoogleAccounts(ctx context.Context, userID int64, tokenOwnerGoogleUserID string, customerIDs []string) error
}

type TokenSource interface {
	Token(ctx context.Context, userID int64) (accessToken string, googleUserID string, err error)
	MarkNeedsConsent(ctx context.Context, userID int64, googleUserID string) error
}

type Client struct {
	http        *http.Client
	devToken    string
	loginMCC    string 
	tokenSource TokenSource
	base        string 
	retries     int
	repo 		linkRepo
}

func New(devToken, loginMCC string, ts TokenSource) *Client {
	return &Client{
		http: &http.Client{ Timeout: 15 * time.Second },
		devToken: devToken,
		loginMCC: strings.ReplaceAll(loginMCC, "-", ""),
		tokenSource: ts,
		base: "https://googleads.googleapis.com/v21",
		retries: 3,
	}
}

func (c *Client) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	backoffs := []time.Duration{500 * time.Millisecond, 2 * time.Second, 5 * time.Second}
	var last error
	for attempt := 0; attempt < c.retries; attempt++ {
		resp, err := c.http.Do(req.WithContext(ctx))
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		last = err
		if attempt < len(backoffs) {
			select {
			case <-time.After(backoffs[attempt]):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	if last == nil {
		last = errors.New("googleads: retries exhausted")
	}
	return nil, last
}

func (c *Client) ListAccessibleCustomers(ctx context.Context, userID int64) ([]string, string, error) {
	accessToken, googleUID, err := c.tokenSource.Token(ctx, userID)
	if err != nil {
		return nil, "", err
	}
	url := c.base + "/customers:listAccessibleCustomers"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("developer-token", c.devToken)
	if c.loginMCC != "" {
		req.Header.Set("login-customer-id", c.loginMCC)
	}

	resp, err := c.do(ctx, req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		_ = c.tokenSource.MarkNeedsConsent(ctx, userID, googleUID)
		return nil, "", fmt.Errorf("unauthorized: re-consent required")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("googleads listAccessible: %s", string(body))
	}
	var out struct {
		ResourceNames []string `json:"resourceNames"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, "", err
	}
	// resourceNames формата "customers/1234567890"
	ids := make([]string, 0, len(out.ResourceNames))
	for _, rn := range out.ResourceNames {
		if i := strings.LastIndex(rn, "/"); i >= 0 {
			ids = append(ids, rn[i+1:])
		}
	}
	return ids, googleUID, nil
}

// Одноразовый синк трат за дату (включительно) по GAQL searchStream
func (c *Client) SyncCostsForDate(ctx context.Context, userID int64, customerID, yyyymmdd string, sink func(adID int64, date string, micros int64) error) error {
	accessToken, googleUID, err := c.tokenSource.Token(ctx, userID)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("%s/customers/%s/googleAds:searchStream", c.base, strings.ReplaceAll(customerID, "-", ""))
	query := `
SELECT ad_group_ad.ad.id, segments.date, metrics.cost_micros
FROM ad_group_ad
WHERE segments.date = '` + yyyymmdd + `'`
	payload := `{"query": ` + jsonQuoted(query) + `}`

	req, _ := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("developer-token", c.devToken)
	if c.loginMCC != "" {
		req.Header.Set("login-customer-id", c.loginMCC)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		_ = c.tokenSource.MarkNeedsConsent(ctx, userID, googleUID)
		return fmt.Errorf("unauthorized: re-consent required")
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("searchStream err: %s", string(b))
	}

	dec := json.NewDecoder(resp.Body)
	for {
		var chunk struct {
			Results []struct {
				AdGroupAd struct {
					Ad struct {
						Id int64 `json:"id,string"`
					} `json:"ad"`
				} `json:"ad_group_ad"`
				Segments struct {
					Date string `json:"date"`
				} `json:"segments"`
				Metrics struct {
					CostMicros int64 `json:"cost_micros,string"`
				} `json:"metrics"`
			} `json:"results"`
		}
		if err := dec.Decode(&chunk); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		for _, r := range chunk.Results {
			if err := sink(r.AdGroupAd.Ad.Id, r.Segments.Date, r.Metrics.CostMicros); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) LinkAccounts(ctx context.Context, userID int64, customerIDs []string) error {
    if len(customerIDs) == 0 {
        return nil
    }
    // Получим google_user_id владельца токена через TokenSource
    _, googleUID, err := c.tokenSource.Token(ctx, userID)
    if err != nil {
        return err
    }
    return c.repo.LinkGoogleAccounts(ctx, userID, googleUID, customerIDs)
}

func jsonQuoted(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
