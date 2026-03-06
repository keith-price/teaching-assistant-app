package drive

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Folder represents a Google Drive folder.
type Folder struct {
	ID   string
	Name string
}

// Client wraps the Google Drive service.
type Client struct {
	srv *drive.Service
}

// NewClient creates a Drive client from an already-authenticated HTTP client.
func NewClient(ctx context.Context, httpClient *http.Client) (*Client, error) {
	srv, err := drive.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %w", err)
	}
	return &Client{srv: srv}, nil
}

// ListFolders returns all subfolders inside a given parent folder ID.
// Use "root" as parentID to list top-level folders.
func (c *Client) ListFolders(ctx context.Context, parentID string) ([]Folder, error) {
	q := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false", parentID)
	r, err := c.srv.Files.List().Q(q).OrderBy("name").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list folders: %w", err)
	}

	var folders []Folder
	for _, f := range r.Files {
		folders = append(folders, Folder{ID: f.Id, Name: f.Name})
	}
	return folders, nil
}

// FindFolder searches for a folder by name within a given parent.
func (c *Client) FindFolder(ctx context.Context, parentID, name string) (*Folder, error) {
	// Escape single quotes in the name to prevent injection/syntax errors in the query
	escapedName := strings.ReplaceAll(name, "'", "\\'")
	q := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and name = '%s' and trashed = false", parentID, escapedName)
	r, err := c.srv.Files.List().Q(q).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to find folder: %w", err)
	}

	if len(r.Files) == 0 {
		return nil, fmt.Errorf("folder not found")
	}

	return &Folder{ID: r.Files[0].Id, Name: r.Files[0].Name}, nil
}

// UploadFile uploads a markdown file to the specified Drive folder and converts it to a Google Doc.
func (c *Client) UploadFile(ctx context.Context, folderID, filename, content string) error {
	// Strip .md extension for Google Docs — they don't need file extensions
	docName := strings.TrimSuffix(filename, ".md")

	f := &drive.File{
		Name:     docName,
		MimeType: "application/vnd.google-apps.document", // Google Docs native format
		Parents:  []string{folderID},
	}
	_, err := c.srv.Files.Create(f).Media(strings.NewReader(content)).Do()
	if err != nil {
		return fmt.Errorf("unable to upload file: %w", err)
	}
	return nil
}

// CreateFolder creates a new subfolder inside the given parent folder and returns it.
func (c *Client) CreateFolder(ctx context.Context, parentID, name string) (*Folder, error) {
	f := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}
	created, err := c.srv.Files.Create(f).Fields("id, name").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create folder: %w", err)
	}
	return &Folder{ID: created.Id, Name: created.Name}, nil
}

