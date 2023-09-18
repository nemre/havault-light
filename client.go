package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const supportedEngineVersion = 2

var (
	ErrKeyNotFound       = errors.New("key not found")
	ErrEngineUnsupported = errors.New("engine unsupported")
)

type Client interface {
	Get(ctx context.Context, key string) (map[string]any, error)
	Set(ctx context.Context, key string, value map[string]any) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

type Config struct {
	Addr   string
	Engine string
	Token  string
}

type h struct {
	c *Config
}

func New(ctx context.Context, c *Config) (Client, error) {
	s := h{c: c}

	res, body, err := s.doReq(ctx, http.MethodGet, fmt.Sprintf("/sys/mounts/%s", c.Engine), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status '%d' received", res.StatusCode)
	}

	d, ok := body["data"].(map[string]any)
	if !ok {
		return nil, ErrEngineUnsupported
	}

	if d["type"] != "kv" {
		return nil, ErrEngineUnsupported
	}

	d, ok = d["options"].(map[string]any)
	if !ok {
		return nil, ErrEngineUnsupported
	}

	if d["version"] != fmt.Sprint(supportedEngineVersion) {
		return nil, ErrEngineUnsupported
	}

	return &s, nil
}

func (h *h) doReq(ctx context.Context, method, path string, body any) (*http.Response, map[string]any, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(body); err != nil {
		return nil, nil, fmt.Errorf("failed to encode request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s/v1%s", h.c.Addr, path), &b)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.c.Token))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to request: %w", err)
	}

	defer res.Body.Close()

	var m map[string]any
	if err = json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return res, m, nil
}

func (h *h) Get(ctx context.Context, key string) (map[string]any, error) {
	res, body, err := h.doReq(ctx, http.MethodGet, fmt.Sprintf("/%s/data/%s", h.c.Engine, key), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request: %w", err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrKeyNotFound
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status '%d' received", res.StatusCode)
	}

	d, ok := body["data"].(map[string]any)
	if !ok {
		return nil, ErrKeyNotFound
	}

	d, ok = d["data"].(map[string]any)
	if !ok {
		return nil, ErrKeyNotFound
	}

	return d, nil
}

func (h *h) Set(ctx context.Context, key string, value map[string]any) error {
	res, _, err := h.doReq(ctx, http.MethodPost, fmt.Sprintf("/%s/data/%s", h.c.Engine, key), map[string]any{"data": value})
	if err != nil {
		return fmt.Errorf("failed to request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status '%d' received", res.StatusCode)
	}

	return nil
}

func (h *h) Delete(ctx context.Context, key string) error {
	res, _, err := h.doReq(ctx, http.MethodDelete, fmt.Sprintf("/%s/data/%s", h.c.Engine, key), nil)
	if err != nil {
		return fmt.Errorf("failed to request: %w", err)
	}

	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected response status '%d' received", res.StatusCode)
	}

	return nil
}

func (h *h) Ping(ctx context.Context) error {
	res, _, err := h.doReq(ctx, http.MethodGet, "/sys/health", nil)
	if err != nil {
		return fmt.Errorf("failed to request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status '%d' received", res.StatusCode)
	}

	return nil
}
