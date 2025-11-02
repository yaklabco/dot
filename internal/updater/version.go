package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jamesainslie/dot/internal/config"
)

// GitHubRelease represents a GitHub release from the API.
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PreRelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
}

// Version represents a semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Raw        string
}

// VersionChecker checks for new versions from GitHub releases.
type VersionChecker struct {
	httpClient *http.Client
	repository string
}

// NewVersionChecker creates a new version checker with default configuration.
func NewVersionChecker(repository string) *VersionChecker {
	return NewVersionCheckerWithConfig(repository, nil)
}

// NewVersionCheckerWithConfig creates a new version checker with explicit network configuration.
func NewVersionCheckerWithConfig(repository string, networkCfg *config.NetworkConfig) *VersionChecker {
	// Use defaults if no config provided
	if networkCfg == nil {
		networkCfg = &config.NetworkConfig{
			Timeout:        10,
			ConnectTimeout: 5,
			TLSTimeout:     5,
		}
	}

	// Create HTTP client with comprehensive timeout configuration
	client := createHTTPClient(networkCfg)

	return &VersionChecker{
		httpClient: client,
		repository: repository,
	}
}

// createHTTPClient creates an HTTP client with comprehensive timeout and proxy configuration.
func createHTTPClient(cfg *config.NetworkConfig) *http.Client {
	// Apply defaults if values are 0
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	connectTimeout := time.Duration(cfg.ConnectTimeout) * time.Second
	if connectTimeout == 0 {
		connectTimeout = 5 * time.Second
	}

	tlsTimeout := time.Duration(cfg.TLSTimeout) * time.Second
	if tlsTimeout == 0 {
		tlsTimeout = 5 * time.Second
	}

	// Create transport with timeout configuration
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   tlsTimeout,
		ResponseHeaderTimeout: 5 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   2,
	}

	// Configure proxy if specified
	if cfg.HTTPProxy != "" || cfg.HTTPSProxy != "" {
		proxyFunc := func(req *http.Request) (*url.URL, error) {
			var proxyURL string
			if req.URL.Scheme == "https" && cfg.HTTPSProxy != "" {
				proxyURL = cfg.HTTPSProxy
			} else if cfg.HTTPProxy != "" {
				proxyURL = cfg.HTTPProxy
			}

			if proxyURL != "" {
				return url.Parse(proxyURL)
			}
			// Fall back to environment
			return http.ProxyFromEnvironment(req)
		}
		transport.Proxy = proxyFunc
	} else {
		// Use environment variables
		transport.Proxy = http.ProxyFromEnvironment
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

// GetLatestVersion fetches the latest release from GitHub.
func (vc *VersionChecker) GetLatestVersion(includePrerelease bool) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", vc.repository)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "dot-updater")

	resp, err := vc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github API returned status %d: %s", resp.StatusCode, string(body))
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("decode releases: %w", err)
	}

	// Find the latest non-draft release
	for _, release := range releases {
		if release.Draft {
			continue
		}
		if release.PreRelease && !includePrerelease {
			continue
		}
		return &release, nil
	}

	return nil, fmt.Errorf("no suitable release found")
}

// ParseVersion parses a version string into a Version struct.
func ParseVersion(versionStr string) (*Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")

	v := &Version{Raw: versionStr}

	// Split on '-' to separate version from pre-release
	parts := strings.SplitN(versionStr, "-", 2)
	if len(parts) == 2 {
		v.PreRelease = parts[1]
	}

	// Parse major.minor.patch
	versionParts := strings.Split(parts[0], ".")
	if len(versionParts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	if _, err := fmt.Sscanf(versionParts[0], "%d", &v.Major); err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}
	if _, err := fmt.Sscanf(versionParts[1], "%d", &v.Minor); err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}
	if _, err := fmt.Sscanf(versionParts[2], "%d", &v.Patch); err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	return v, nil
}

// IsNewerThan returns true if v is newer than other.
func (v *Version) IsNewerThan(other *Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch > other.Patch
	}

	// If versions are equal, consider pre-release
	// Special case: development versions like "2-g14ba5af" indicate commits ahead
	// These should be considered newer than the base release
	if v.PreRelease != "" && strings.Contains(v.PreRelease, "-g") {
		// Development build (e.g., "2-g14ba5af" means 2 commits ahead of release)
		// Consider this newer than the base release
		if other.PreRelease == "" {
			return true
		}
	}

	// A release version is newer than a pre-release version
	if v.PreRelease == "" && other.PreRelease != "" {
		// But not if other is a development build ahead of this release
		if strings.Contains(other.PreRelease, "-g") {
			return false
		}
		return true
	}
	if v.PreRelease != "" && other.PreRelease == "" {
		// If v is a development build, we already handled it above
		return false
	}

	// Both are pre-releases or both are releases - consider equal
	return false
}

// String returns the version as a string.
func (v *Version) String() string {
	if v.PreRelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.PreRelease)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// CheckForUpdate checks if there's a newer version available.
func (vc *VersionChecker) CheckForUpdate(currentVersion string, includePrerelease bool) (newVersion *GitHubRelease, hasUpdate bool, err error) {
	current, err := ParseVersion(currentVersion)
	if err != nil {
		return nil, false, fmt.Errorf("parse current version: %w", err)
	}

	latest, err := vc.GetLatestVersion(includePrerelease)
	if err != nil {
		return nil, false, fmt.Errorf("get latest version: %w", err)
	}

	latestVersion, err := ParseVersion(latest.TagName)
	if err != nil {
		return nil, false, fmt.Errorf("parse latest version: %w", err)
	}

	if latestVersion.IsNewerThan(current) {
		return latest, true, nil
	}

	return latest, false, nil
}
