# Implement Markdown to Google Docs Conversion

The Google Drive API does not natively convert Markdown into formatted Google Docs. While setting the `MimeType` to `application/vnd.google-apps.document` creates a Google Doc, the raw Markdown text is simply inserted without parsing the formatting (bold, headers, lists, etc).

The official and most robust way to achieve rich text conversion via the Drive API is to upload **HTML** and instruct Drive to convert it.

## Proposed Changes

### 1. New Dependency

We need a lightweight, robust Markdown-to-HTML converter for Go. I propose adding `github.com/gomarkdown/markdown`.

```bash
go get github.com/gomarkdown/markdown
```

### 2. Update `internal/drive/drive.go`

I will modify the `UploadFile` function to accept the markdown content, convert it to HTML in memory, and then upload the HTML byte stream to Google Drive.

#### [MODIFY] drive.go

- Import `github.com/gomarkdown/markdown` and `github.com/gomarkdown/markdown/html`.
- In `UploadFile`, convert the `content` string (which is markdown) into HTML bytes.
- When calling `c.srv.Files.Create(f).Media(...)`, pass the HTML bytes instead of the raw markdown string.
- (Optional but recommended) explicitly set the media content type to `text/html` during the `Media()` call to ensure the Drive API recognizes the source format for conversion:
  ```go
  import "google.golang.org/api/googleapi"
  // ...
  .Media(bytes.NewReader(htmlContent), googleapi.ContentType("text/html"))
  ```

### 3. Update Tests

#### [MODIFY] drive_test.go

- No structural changes needed to the mock interface, but we will add a test case verifying that `UploadFile` handles the conversion gracefully (the mock itself won't call the real API, but we can verify it doesn't panic on the new HTML conversion logic).

## Verification Plan

### Automated Tests

1. Run `go test ./internal/drive/ -v` to ensure the mock and logic still pass.
2. Run `go build ./...` to verify the new dependency compiles cleanly across the project.

### Manual Verification (User)

I will ask Keith to:

1. Run the app (`go run cmd/app/main.go`).
2. Generate a new worksheet for any text.
3. Select a Google Drive folder to upload to.
4. Open the newly created Google Docs in the browser and verify that headers (H1, H2), bold text, and lists are natively formatted as rich text, not raw markdown symbols (`**`, `#`, etc).

---

## Junior Dev Handoff

The Senior Dev (Critic) has investigated the issue and formulated the plan above.

**Junior Dev Action Required:**

1. Please review this plan.
2. Implement the changes to `drive.go` and `drive_test.go`, including adding the `gomarkdown` dependency.
3. Verify it works locally.
4. Add a handoff note to `PLAN.md` when completed.
5. Keith will then pass your changes back to me for the final Senior Dev review.
