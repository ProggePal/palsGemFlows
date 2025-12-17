package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// EnvRecipesBaseURL configures the remote recipe catalog.
	// Example: https://raw.githubusercontent.com/MYORG/automation-platform/main/recipes/
	EnvRecipesBaseURL = "PALSGEMFLOWS_RECIPES_BASE_URL"

	// DefaultRecipesBaseURL is the default remote catalog for this project.
	// Override via EnvRecipesBaseURL or the --recipes-base-url flag.
	DefaultRecipesBaseURL = "https://raw.githubusercontent.com/ProggePal/palsGemFlows/main/workflows/"

	defaultGitHubRepo = "ProggePal/palsGemFlows"
	defaultGitHubRef  = "main"
	defaultGitHubDir  = "workflows"

	defaultCacheTTL = 1 * time.Hour
)

var ErrRemoteNotConfigured = errors.New("remote recipes are not configured")

type Source string

const (
	SourceRemote   Source = "remote"
	SourceLocalDev Source = "local_dev"
)

type Result struct {
	Data       []byte
	Source     Source
	RecipeName string // for analytics (e.g. "marketing/blog_post" or local path)
	URL        string // populated for remote
}

type Options struct {
	BaseURL  string
	CacheTTL time.Duration
	Client   *http.Client
}

type RemoteListOptions struct {
	Client *http.Client
	Ref    string
}

// ListRemoteRecipeKeys lists available workflow keys from the default GitHub repo.
// This is best-effort and intended for interactive UX (not required for running by name).
func ListRemoteRecipeKeys(ctx context.Context, opts RemoteListOptions) ([]string, error) {
	ref := strings.TrimSpace(opts.Ref)
	if ref == "" {
		ref = defaultGitHubRef
	}

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s", defaultGitHubRepo, defaultGitHubDir, ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// GitHub API expects a User-Agent.
	req.Header.Set("User-Agent", "pals-gemflows")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list remote recipes failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	var entries []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}

	seen := map[string]struct{}{}
	var out []string
	for _, e := range entries {
		if e.Type != "file" {
			continue
		}
		low := strings.ToLower(e.Name)
		if !strings.HasSuffix(low, ".yaml") && !strings.HasSuffix(low, ".yml") {
			continue
		}
		key := stripYAMLExt(e.Name)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out, nil
}

func GetRecipeData(ctx context.Context, nameOrPath string, opts Options) (Result, error) {
	nameOrPath = strings.TrimSpace(nameOrPath)
	if nameOrPath == "" {
		return Result{}, errors.New("recipe name or path is required")
	}

	// 1) Local path mode (dev): if the OS can stat it, treat it as a file.
	if st, err := os.Stat(nameOrPath); err == nil && !st.IsDir() {
		b, err := os.ReadFile(nameOrPath)
		if err != nil {
			return Result{}, fmt.Errorf("read local recipe %s: %w", nameOrPath, err)
		}
		return Result{Data: b, Source: SourceLocalDev, RecipeName: nameOrPath}, nil
	}

	// 2) Remote catalog mode.
	baseURL := strings.TrimSpace(opts.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv(EnvRecipesBaseURL))
	}
	if baseURL == "" {
		baseURL = DefaultRecipesBaseURL
	}
	if baseURL == "" {
		return Result{}, fmt.Errorf("%w: set %s to your GitHub Raw recipes base URL", ErrRemoteNotConfigured, EnvRecipesBaseURL)
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	recipePath := normalizeRemoteRecipePath(nameOrPath)
	url := baseURL + recipePath

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	ttl := opts.CacheTTL
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}

	cachePath := cacheFilePath(url)
	if b, ok := readFreshCache(cachePath, ttl); ok {
		return Result{Data: b, Source: SourceRemote, RecipeName: stripYAMLExt(recipePath), URL: url}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		// If the network failed but we have any cache, return it as a fallback.
		if b, ok := readAnyCache(cachePath); ok {
			return Result{Data: b, Source: SourceRemote, RecipeName: stripYAMLExt(recipePath), URL: url}, nil
		}
		return Result{}, fmt.Errorf("fetch remote recipe: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if b, ok := readAnyCache(cachePath); ok {
			return Result{Data: b, Source: SourceRemote, RecipeName: stripYAMLExt(recipePath), URL: url}, nil
		}
		return Result{}, fmt.Errorf("recipe not found in catalog (HTTP %d): %s", resp.StatusCode, url)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("read remote recipe: %w", err)
	}

	_ = writeCache(cachePath, b)
	return Result{Data: b, Source: SourceRemote, RecipeName: stripYAMLExt(recipePath), URL: url}, nil
}

func normalizeRemoteRecipePath(name string) string {
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSpace(name)
	if !strings.HasSuffix(strings.ToLower(name), ".yaml") && !strings.HasSuffix(strings.ToLower(name), ".yml") {
		name += ".yaml"
	}
	return name
}

func stripYAMLExt(path string) string {
	low := strings.ToLower(path)
	switch {
	case strings.HasSuffix(low, ".yaml"):
		return path[:len(path)-len(".yaml")]
	case strings.HasSuffix(low, ".yml"):
		return path[:len(path)-len(".yml")]
	default:
		return path
	}
}

func cacheFilePath(url string) string {
	dir, err := os.UserCacheDir()
	if err != nil || dir == "" {
		// Fallback to current directory if we can't find a cache dir.
		dir = "."
	}
	h := sha256.Sum256([]byte(url))
	name := hex.EncodeToString(h[:]) + ".yaml"
	return filepath.Join(dir, "pals-gemflows", "recipes", name)
}

func readFreshCache(path string, ttl time.Duration) ([]byte, bool) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, false
	}
	if ttl > 0 && time.Since(st.ModTime()) > ttl {
		return nil, false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return b, true
}

func readAnyCache(path string) ([]byte, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return b, true
}

func writeCache(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// best-effort
	return os.WriteFile(path, data, 0o644)
}
