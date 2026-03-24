package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	NYT    NYTConfig    `toml:"nyt"`
	Print  PrintConfig  `toml:"print"`
	Upload UploadConfig `toml:"upload"`
}

type NYTConfig struct {
	Email    string `toml:"email"`
	Password string `toml:"password"`
	Token    string `toml:"token"`
}

type PrintConfig struct {
	Printer string `toml:"printer"`
}

type UploadConfig struct {
	Host       string `toml:"host"`
	KeyPath    string `toml:"key_path"`
	RemotePath string `toml:"remote_path"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	// 1. Config file as base
	if path, err := configFilePath(); err == nil {
		if _, err := os.Stat(path); err == nil {
			if _, err := toml.DecodeFile(path, cfg); err != nil {
				return nil, fmt.Errorf("parsing config file: %w", err)
			}
		}
	}

	// 2. macOS Keychain overrides config file (no-op on Linux)
	if email, err := keychainGet("email"); err == nil && email != "" {
		cfg.NYT.Email = email
	}
	if password, err := keychainGet("password"); err == nil && password != "" {
		cfg.NYT.Password = password
	}

	// 3. Environment variables take highest precedence
	if v := os.Getenv("NYT_EMAIL"); v != "" {
		cfg.NYT.Email = v
	}
	if v := os.Getenv("NYT_PASSWORD"); v != "" {
		cfg.NYT.Password = v
	}
	if v := os.Getenv("NYT_S"); v != "" {
		cfg.NYT.Token = v
	}
	if v := os.Getenv("PRINTER"); v != "" {
		cfg.Print.Printer = v
	}
	if v := os.Getenv("UPLOAD_HOST"); v != "" {
		cfg.Upload.Host = v
	}
	if v := os.Getenv("UPLOAD_KEY_PATH"); v != "" {
		cfg.Upload.KeyPath = v
	}
	if v := os.Getenv("UPLOAD_REMOTE_PATH"); v != "" {
		cfg.Upload.RemotePath = v
	}

	// 4. Resolve any remaining op:// references via the 1Password CLI
	var err error
	if cfg.NYT.Email, err = resolveSecret(cfg.NYT.Email); err != nil {
		return nil, fmt.Errorf("resolving email: %w", err)
	}
	if cfg.NYT.Password, err = resolveSecret(cfg.NYT.Password); err != nil {
		return nil, fmt.Errorf("resolving password: %w", err)
	}
	if cfg.NYT.Token, err = resolveSecret(cfg.NYT.Token); err != nil {
		return nil, fmt.Errorf("resolving token: %w", err)
	}

	if cfg.NYT.Token == "" && (cfg.NYT.Email == "" || cfg.NYT.Password == "") {
		return nil, fmt.Errorf(`NYT credentials not found.

Set them up with one of these methods:

  Option 1 - NYT-S cookie (recommended for Docker):
    NYT_S=<value of NYT-S cookie from your browser>
    PRINTER=ipp://192.168.1.x/ipp/print   (optional)

  Option 2 - Email/password environment variables:
    NYT_EMAIL=you@example.com
    NYT_PASSWORD=yourpassword
    PRINTER=ipp://192.168.1.x/ipp/print   (optional)

  Option 3 - 1Password (op:// references in config file at %s):
    [nyt]
    email    = "op://Personal/New York Times/username"
    password = "op://Personal/New York Times/password"

  Option 4 - macOS Keychain:
    security add-generic-password -s puzzle-printer-nyt -a email -w "you@example.com"
    security add-generic-password -s puzzle-printer-nyt -a password -w "yourpassword"

  Option 5 - Plaintext config file at %s:
    [nyt]
    email    = "you@example.com"
    password = "yourpassword"`, mustConfigFilePath(), mustConfigFilePath())
	}

	return cfg, nil
}

// resolveSecret expands an op:// secret reference via the 1Password CLI.
// Non-op:// values are returned as-is.
func resolveSecret(value string) (string, error) {
	if !strings.HasPrefix(value, "op://") {
		return value, nil
	}
	out, err := exec.Command("op", "read", "--no-newline", value).Output()
	if err != nil {
		return "", fmt.Errorf("op read %q: %w", value, err)
	}
	return string(out), nil
}

func keychainGet(account string) (string, error) {
	out, err := exec.Command("security", "find-generic-password",
		"-s", "puzzle-printer-nyt",
		"-a", account,
		"-w",
	).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "puzzle-printer", "config.toml"), nil
}

func mustConfigFilePath() string {
	p, _ := configFilePath()
	return p
}
