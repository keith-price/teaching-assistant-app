package drive

import (
	"context"
	"strings"
	"testing"

	"github.com/gomarkdown/markdown"
)

// DriveUploader allows mocking of the Drive service in tests.
type DriveUploader interface {
	ListFolders(ctx context.Context, parentID string) ([]Folder, error)
	FindFolder(ctx context.Context, parentID, name string) (*Folder, error)
	UploadFile(ctx context.Context, folderID, filename, content string) error
	CreateFolder(ctx context.Context, parentID, name string) (*Folder, error)
}

type MockDriveClient struct {
	folders []Folder
	files   map[string]string // filename to content
}

func (m *MockDriveClient) ListFolders(ctx context.Context, parentID string) ([]Folder, error) {
	return m.folders, nil
}

func (m *MockDriveClient) FindFolder(ctx context.Context, parentID, name string) (*Folder, error) {
	for _, f := range m.folders {
		if f.Name == name {
			return &f, nil
		}
	}
	return nil, nil // Or an error depending on how we handle it
}

func (m *MockDriveClient) UploadFile(ctx context.Context, folderID, filename, content string) error {
	m.files[filename] = content
	return nil
}

func (m *MockDriveClient) CreateFolder(ctx context.Context, parentID, name string) (*Folder, error) {
	f := Folder{ID: "new-" + name, Name: name}
	m.folders = append(m.folders, f)
	return &f, nil
}

func TestMockDriveClient(t *testing.T) {
	mock := &MockDriveClient{
		folders: []Folder{{ID: "1", Name: "folder1"}, {ID: "2", Name: "folder2"}},
		files:   make(map[string]string),
	}

	ctx := context.Background()
	folders, err := mock.ListFolders(ctx, "root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(folders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(folders))
	}

	err = mock.UploadFile(ctx, "1", "test.md", "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.files["test.md"] != "content" {
		t.Errorf("file not uploaded correctly")
	}
}

func TestMockCreateFolder(t *testing.T) {
	mock := &MockDriveClient{
		folders: []Folder{},
		files:   make(map[string]string),
	}

	ctx := context.Background()
	folder, err := mock.CreateFolder(ctx, "root", "test_folder")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if folder.Name != "test_folder" || folder.ID != "new-test_folder" {
		t.Errorf("folder created incorrectly: %+v", folder)
	}

	if len(mock.folders) != 1 {
		t.Errorf("expected 1 folder, got %d", len(mock.folders))
	}
}

func TestMarkdownConversion(t *testing.T) {
	content := "# Generic Title\n\n## Section 1\n\n- **Bold Item**: description\n- ________: a gap-fill item"
	htmlBytes := markdown.ToHTML([]byte(content), nil, nil)
	html := string(htmlBytes)

	if !strings.Contains(html, "<h1>Generic Title</h1>") {
		t.Errorf("Expected H1 tag for title, got: %s", html)
	}
	if !strings.Contains(html, "<h2>Section 1</h2>") {
		t.Errorf("Expected H2 tag for section, got: %s", html)
	}
	if !strings.Contains(html, "<strong>Bold Item</strong>") {
		t.Errorf("Expected strong tag for bold item, got: %s", html)
	}
	if !strings.Contains(html, "________:") {
		t.Errorf("Expected gap-fill underscores to be preserved, got: %s", html)
	}
}
