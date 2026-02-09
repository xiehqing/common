package update

import "context"

type Client interface {
	Latest(ctx context.Context) (*Release, error)
}

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type github struct{}
