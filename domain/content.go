package domain

// Content represents one piece of content to handle
type Content struct {
	IsDirectory  bool
	IsCompressed bool
	FullPath     string
}
