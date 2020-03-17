package domain

// Content represents one piece of content to handle
type Content struct {
	IsDirectory     bool
	ItemName        string
	FullPath        string
	DestinationPath string
	MediaContent    []string
}
