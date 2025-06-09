package userclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// User — структура, ожидаемая от User-service
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"createdAt"`
}

// Client — HTTP клиент для связи с User-service
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New создаёт новый клиент
func New(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetByEmail — отправляет GET-запрос на /api/users/by-email?email=...
func (c *Client) GetByEmail(ctx context.Context, email string) (User, error) {
	url := fmt.Sprintf("%s/api/users/by-email?email=%s", c.BaseURL, email)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return User{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return User{}, fmt.Errorf("decode response: %w", err)
	}
	return user, nil
}
