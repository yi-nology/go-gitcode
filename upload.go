package gitcode

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

// multipartBody builds a multipart form with a single "file" field and
// returns the body together with its Content-Type. The body is a plain
// []byte so it can be replayed across retries (doRequestWithRetry resets
// the body via req.GetBody, which http.NewRequestWithContext wires for
// *bytes.Reader bodies).
func multipartBody(filename string, reader io.Reader) (body []byte, contentType string, err error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, reader); err != nil {
		return nil, "", fmt.Errorf("copy file content: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("close multipart writer: %w", err)
	}
	return buf.Bytes(), writer.FormDataContentType(), nil
}

// UploadFileReader uploads a file using multipart form data.
// This is the proper way to upload file contents (as opposed to UploadFile which takes a path).
//
// POST /repos/{owner}/{repo}/file/upload
//
// Uploads run through the shared request pipeline (doBufferedRequestAt):
// request/response hooks fire and the retry policy applies, exactly like
// every other request method (issue #9).
func (c *Client) UploadFileReader(ctx context.Context, owner, repo, filename string, reader io.Reader) (*FileUploadResult, error) {
	if err := validateOwnerRepo(owner, repo); err != nil {
		return nil, err
	}

	body, contentType, err := multipartBody(filename, reader)
	if err != nil {
		return nil, err
	}

	var result FileUploadResult
	base, _ := c.requestSnapshot()
	if err := c.doBufferedRequestAt(ctx, base, http.MethodPost,
		pathf("/repos/%s/%s/file/upload", owner, repo), contentType, body, &result, c.jsonBodyLimit()); err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadImageReader uploads an image using multipart form data.
//
// POST /repos/{owner}/{repo}/img/upload
//
// Uploads run through the shared request pipeline (doBufferedRequestAt):
// request/response hooks fire and the retry policy applies, exactly like
// every other request method (issue #9).
func (c *Client) UploadImageReader(ctx context.Context, owner, repo, filename string, reader io.Reader) (*FileUploadResult, error) {
	if err := validateOwnerRepo(owner, repo); err != nil {
		return nil, err
	}

	body, contentType, err := multipartBody(filename, reader)
	if err != nil {
		return nil, err
	}

	var result FileUploadResult
	base, _ := c.requestSnapshot()
	if err := c.doBufferedRequestAt(ctx, base, http.MethodPost,
		pathf("/repos/%s/%s/img/upload", owner, repo), contentType, body, &result, c.jsonBodyLimit()); err != nil {
		return nil, err
	}
	return &result, nil
}

// UploadFileBytes uploads file bytes using multipart form data.
func (c *Client) UploadFileBytes(ctx context.Context, owner, repo, filename string, data []byte) (*FileUploadResult, error) {
	return c.UploadFileReader(ctx, owner, repo, filename, bytes.NewReader(data))
}

// UploadFileWithPath uploads a file from a local path using multipart form data.
func (c *Client) UploadFileWithPath(ctx context.Context, owner, repo, localPath string) (*FileUploadResult, error) {
	// This is a convenience wrapper - the caller should read the file themselves
	// and pass an io.Reader. This method is kept for backward compatibility.
	filename := filepath.Base(localPath)
	return c.UploadFile(ctx, owner, repo, filename)
}
