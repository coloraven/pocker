package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pocker/pocker/internal/config"
)

func TestLoadSaveGetSet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("POCKER_CONFIG_DIR", dir)

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Set("http_proxy", "http://127.0.0.1:7890"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.Set("HTTPS_PROXY", "http://127.0.0.1:7890"); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, "config.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.Get("http_proxy")
	if err != nil || got != "http://127.0.0.1:7890" {
		t.Fatalf("http_proxy=%q err=%v", got, err)
	}
	got, err = loaded.Get("https_proxy")
	if err != nil || got != "http://127.0.0.1:7890" {
		t.Fatalf("https_proxy=%q err=%v", got, err)
	}

	if err := loaded.Unset("http_proxy"); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(loaded); err != nil {
		t.Fatal(err)
	}
	again, _ := config.Load()
	got, _ = again.Get("http_proxy")
	if got != "" {
		t.Fatalf("expected unset, got %q", got)
	}
}

func TestProxyEnvPrefersOS(t *testing.T) {
	cfg := &config.Config{HTTPProxy: "http://from-config:1"}
	t.Setenv("HTTP_PROXY", "http://from-env:2")
	extra := cfg.ProxyEnv()
	for _, e := range extra {
		if e == "HTTP_PROXY=http://from-config:1" || e == "http_proxy=http://from-config:1" {
			t.Fatalf("config should not override existing env, got %v", extra)
		}
	}
}

func TestProxyEnvSOCKSPromotesAllProxy(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("http_proxy", "")
	t.Setenv("ALL_PROXY", "")
	t.Setenv("all_proxy", "")

	cfg := &config.Config{HTTPProxy: "socks5://127.0.0.1:1080"}
	extra := cfg.ProxyEnv()
	want := map[string]bool{
		"HTTP_PROXY=socks5://127.0.0.1:1080": false,
		"http_proxy=socks5://127.0.0.1:1080": false,
		"ALL_PROXY=socks5://127.0.0.1:1080":  false,
		"all_proxy=socks5://127.0.0.1:1080":  false,
	}
	for _, e := range extra {
		if _, ok := want[e]; ok {
			want[e] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Fatalf("missing %s in %v", k, extra)
		}
	}
}

func TestAllProxyKey(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("POCKER_CONFIG_DIR", dir)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Set("all_proxy", "socks5h://127.0.0.1:1080"); err != nil {
		t.Fatal(err)
	}
	if err := config.Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.Get("all_proxy")
	if err != nil || got != "socks5h://127.0.0.1:1080" {
		t.Fatalf("all_proxy=%q err=%v", got, err)
	}
}
