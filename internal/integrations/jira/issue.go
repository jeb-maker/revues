package jira

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var issueKeyPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]+-\d+$`)

// ErrInvalidIssueReference is returned when the input is not a Jira key or browse URL.
var ErrInvalidIssueReference = errors.New("clé ou URL Jira invalide")

// ParseIssueReference extracts a Jira issue key from a key or browse URL.
func ParseIssueReference(raw string) (key string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("référence issue Jira requise")
	}

	candidate := strings.ToUpper(raw)
	if issueKeyPattern.MatchString(candidate) {
		return candidate, nil
	}

	u, parseErr := url.Parse(raw)
	if parseErr != nil || u.Scheme == "" || u.Host == "" {
		return "", ErrInvalidIssueReference
	}

	path := strings.TrimSuffix(u.Path, "/")
	if idx := strings.LastIndex(path, "/browse/"); idx >= 0 {
		candidate = strings.ToUpper(path[idx+len("/browse/"):])
	} else {
		segments := strings.Split(strings.Trim(path, "/"), "/")
		if len(segments) == 0 {
			return "", ErrInvalidIssueReference
		}
		candidate = strings.ToUpper(segments[len(segments)-1])
	}

	if q := strings.Index(candidate, "?"); q >= 0 {
		candidate = candidate[:q]
	}
	if !issueKeyPattern.MatchString(candidate) {
		return "", ErrInvalidIssueReference
	}

	return candidate, nil
}

// BrowseURL builds a Jira browse URL for an issue key.
func BrowseURL(baseURL, key string) string {
	return NormalizeBaseURL(baseURL) + "/browse/" + strings.ToUpper(strings.TrimSpace(key))
}

// ValidateBrowseURL ensures the URL belongs to the configured Jira host.
func ValidateBrowseURL(cfg Config, issueURL string) error {
	issueURL = strings.TrimSpace(issueURL)
	if issueURL == "" {
		return errors.New("URL Jira requise")
	}

	u, err := url.Parse(issueURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("URL Jira invalide")
	}

	switch u.Scheme {
	case "https":
	case "http":
		host := u.Hostname()
		if host != "localhost" && host != "127.0.0.1" {
			return errors.New("URL Jira doit utiliser HTTPS")
		}
	default:
		return errors.New("URL Jira doit utiliser HTTPS")
	}

	base, err := url.Parse(NormalizeBaseURL(cfg.BaseURL))
	if err != nil || base.Host == "" {
		return errors.New("URL Jira de configuration invalide")
	}
	if !strings.EqualFold(u.Hostname(), base.Hostname()) {
		return errors.New("URL Jira ne correspond pas à l'instance configurée")
	}

	return nil
}
