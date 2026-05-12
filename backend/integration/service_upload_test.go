//go:build integration

package integration

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wetalk/service"
)

func TestServiceUpload_Image(t *testing.T) {
	oss := &stubOSS{}
	svc := service.NewUploadService(oss)

	resp, err := svc.Upload(context.Background(), 1, "image",
		bytes.NewReader([]byte("fake-image-data")),
		"photo.jpg", 1024, "image/jpeg")
	require.NoError(t, err)
	assert.Equal(t, "photo.jpg", resp.FileName)
	assert.Equal(t, int64(1024), resp.FileSize)
	assert.Contains(t, resp.URL, "uploads/image/1/")
	assert.Contains(t, resp.URL, "test-bucket.oss-cn-test.aliyuncs.com")
}

func TestServiceUpload_File(t *testing.T) {
	oss := &stubOSS{}
	svc := service.NewUploadService(oss)

	resp, err := svc.Upload(context.Background(), 1, "file",
		bytes.NewReader([]byte("fake-file-data")),
		"document.pdf", 5000, "application/pdf")
	require.NoError(t, err)
	assert.Equal(t, "document.pdf", resp.FileName)
	assert.Contains(t, resp.URL, "uploads/file/1/")
}

func TestServiceUpload_InvalidType(t *testing.T) {
	oss := &stubOSS{}
	svc := service.NewUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, "video",
		nil, "video.mp4", 100, "video/mp4")
	require.Error(t, err)
}

func TestServiceUpload_ImageTooLarge(t *testing.T) {
	oss := &stubOSS{}
	svc := service.NewUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, "image",
		nil, "large.jpg", 11*1024*1024, "image/jpeg")
	require.Error(t, err)
}

func TestServiceUpload_FileTooLarge(t *testing.T) {
	oss := &stubOSS{}
	svc := service.NewUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, "file",
		nil, "large.pdf", 21*1024*1024, "application/pdf")
	require.Error(t, err)
}

func TestServiceUpload_NotImageMime(t *testing.T) {
	oss := &stubOSS{}
	svc := service.NewUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, "image",
		nil, "fake.jpg", 100, "application/pdf")
	require.Error(t, err)
}

func TestServiceUpload_VerifyOSSCall(t *testing.T) {
	oss := &stubOSS{}
	svc := service.NewUploadService(oss)

	_, err := svc.Upload(context.Background(), 42, "file",
		bytes.NewReader([]byte("hello")),
		"test.txt", 10, "text/plain")
	require.NoError(t, err)

	// Verify OSS was called
	require.Len(t, oss.calls, 1)
	assert.Contains(t, oss.calls[0].Key, "uploads/file/42/")
	assert.Equal(t, int64(10), oss.calls[0].ContentLength)
	assert.Equal(t, "text/plain", oss.calls[0].ContentType)
}
