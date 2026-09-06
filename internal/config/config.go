package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	KeyHTTPProxy  = "http_proxy"
	KeyHTTPSProxy = "https_proxy"
	KeyAllProxy   = "all_proxy"
	KeyNoProxy    = "no_proxy"
	KeySkopeoPath = "skopeo_path"
)

var knownKeys = map[string]bool{
	KeyHTTPProxy:  true,
	KeyHTTPSProxy: true,
	KeyAllProxy:   true,
	KeyNoProxy:    true,
	KeySkopeoPath: true,
}

// Config holds persistent pocker settings.
type Config struct {
	HTTPProxy  string `json:"http_proxy,omitempty"`
	HTTPSProxy string `json:"https_proxy,omitempty"`
	AllProxy   string `json:"all_proxy,omitempty"`
	NoProxy    string `json:"no_proxy,omitempty"`
	SkopeoPath string `json:"skopeo_path,omitempty"`
}

// Path returns the config file path for the current OS.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Dir returns the config directory.
func Dir() (string, error) {
	if override := os.Getenv("POCKER_CONFIG_DIR"); override != "" {
		return override, nil
	}
	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("APPDATA")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(base, "pocker"), nil
	default:
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, "pocker"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", "pocker"), nil
	}
}

// Load reads config from disk. Missing file yields empty config.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &cfg, nil
}

// Save writes config to disk.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path, err := Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

// NormalizeKey lowercases and validates a config key.
func NormalizeKey(key string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(key))
	if !knownKeys[k] {
		return "", fmt.Errorf("unknown config key %q (supported: http_proxy, https_proxy, all_proxy, no_proxy, skopeo_path)", key)
	}
	return k, nil
}

// Get returns a config value by key.
func (c *Config) Get(key string) (string, error) {
	k, err := NormalizeKey(key)
	if err != nil {
		return "", err
	}
	switch k {
	case KeyHTTPProxy:
		return c.HTTPProxy, nil
	case KeyHTTPSProxy:
		return c.HTTPSProxy, nil
	case KeyAllProxy:
		return c.AllProxy, nil
	case KeyNoProxy:
		return c.NoProxy, nil
	case KeySkopeoPath:
		return c.SkopeoPath, nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
}

// Set sets a config value by key.
func (c *Config) Set(key, value string) error {
	k, err := NormalizeKey(key)
	if err != nil {
		return err
	}
	switch k {
	case KeyHTTPProxy:
		c.HTTPProxy = value
	case KeyHTTPSProxy:
		c.HTTPSProxy = value
	case KeyAllProxy:
		c.AllProxy = value
	case KeyNoProxy:
		c.NoProxy = value
	case KeySkopeoPath:
		c.SkopeoPath = value
	}
	return nil
}

// Unset clears a config value by key.
func (c *Config) Unset(key string) error {
	return c.Set(key, "")
}

// Map returns all non-empty settings.
func (c *Config) Map() map[string]string {
	out := map[string]string{}
	if c.HTTPProxy != "" {
		out[KeyHTTPProxy] = c.HTTPProxy
	}
	if c.HTTPSProxy != "" {
		out[KeyHTTPSProxy] = c.HTTPSProxy
	}
	if c.AllProxy != "" {
		out[KeyAllProxy] = c.AllProxy
	}
	if c.NoProxy != "" {
		out[KeyNoProxy] = c.NoProxy
	}
	if c.SkopeoPath != "" {
		out[KeySkopeoPath] = c.SkopeoPath
	}
	return out
}

func isSOCKS(url string) bool {
	u := strings.ToLower(strings.TrimSpace(url))
	return strings.HasPrefix(u, "socks5://") ||
		strings.HasPrefix(u, "socks5h://") ||
		strings.HasPrefix(u, "socks4://") ||
		strings.HasPrefix(u, "socks4a://") ||
		strings.HasPrefix(u, "socks://")
}

func envAlreadySet(keys ...string) bool {
	for _, k := range keys {
		if os.Getenv(k) != "" {
			return true
		}
	}
	return false
}

func appendEnvPair(extra []string, keys []string, value string) []string {
	if value == "" || envAlreadySet(keys...) {
		return extra
	}
	for _, k := range keys {
		extra = append(extra, k+"="+value)
	}
	return extra
}

// ProxyEnv returns environment assignments for skopeo, preferring existing OS env.
// SOCKS proxies (socks5://, socks5h://, …) are also exported as ALL_PROXY/all_proxy
// because containers/image and Go's http client resolve SOCKS via ALL_PROXY.
func (c *Config) ProxyEnv() []string {
	var extra []string

	extra = appendEnvPair(extra, []string{"HTTP_PROXY", "http_proxy"}, c.HTTPProxy)
	extra = appendEnvPair(extra, []string{"HTTPS_PROXY", "https_proxy"}, c.HTTPSProxy)
	extra = appendEnvPair(extra, []string{"NO_PROXY", "no_proxy"}, c.NoProxy)

	all := c.AllProxy
	if all == "" {
		// Promote SOCKS URLs from http(s)_proxy into ALL_PROXY when all_proxy unset.
		switch {
		case isSOCKS(c.HTTPProxy):
			all = c.HTTPProxy
		case isSOCKS(c.HTTPSProxy):
			all = c.HTTPSProxy
		}
	}
	extra = appendEnvPair(extra, []string{"ALL_PROXY", "all_proxy"}, all)

	return extra
}
