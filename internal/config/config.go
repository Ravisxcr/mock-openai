// Package config loads server configuration from the environment,
// optionally seeded by a .env file for local/docker runs.
package config

import (
	"bufio"
	"os"
	"strings"
)

// LoadDotEnv reads simple KEY=VALUE pairs from the given file (if it exists)
// into the process environment. Blank lines and lines starting with '#' are
// skipped. Real environment variables always win — a value already set via
// os.Setenv/the shell is never overwritten by the file.
func LoadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// APIKey returns the API_KEY configured via .env/environment, or "" if unset.
func APIKey() string {
	return os.Getenv("API_KEY")
}
