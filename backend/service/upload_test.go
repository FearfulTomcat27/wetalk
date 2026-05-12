package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"wetalk/model"
	"wetalk/types"
)

func newTestUploadService(oss *MockOSS) *UploadService {
	return NewUploadService(oss)
}

func TestUploadService_InvalidType(t *testing.T) {
	oss := new(MockOSS)
	svc := newTestUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, "video", nil, "test.mp4", 100, "video/mp4")
	assert.ErrorIs(t, err, types.ErrInvalidUploadType)
}

func TestUploadService_ImageNotImage(t *testing.T) {
	oss := new(MockOSS)
	svc := newTestUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, model.ContentTypeImage, nil, "test.txt", 100, "text/plain")
	assert.ErrorIs(t, err, types.ErrNotImage)
}

func TestUploadService_ImageTooLarge(t *testing.T) {
	oss := new(MockOSS)
	svc := newTestUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, model.ContentTypeImage, nil, "photo.jpg", 11*1024*1024, "image/jpeg")
	assert.ErrorIs(t, err, types.ErrImageTooLarge)
}

func TestUploadService_FileTooLarge(t *testing.T) {
	oss := new(MockOSS)
	svc := newTestUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, model.ContentTypeFile, nil, "doc.pdf", 21*1024*1024, "application/pdf")
	assert.ErrorIs(t, err, types.ErrFileTooLarge)
}

func TestUploadService_FileSuccess(t *testing.T) {
	oss := new(MockOSS)
	oss.On("PutObject", mock.Anything, mock.MatchedBy(func(key string) bool {
		return strings.Contains(key, "uploads/file/1/")
	}), mock.Anything, "application/pdf", int64(4)).
		Return("https://bucket.oss-cn-hangzhou.aliyuncs.com/uploads/file/1/test.pdf", nil)

	svc := newTestUploadService(oss)

	resp, err := svc.Upload(context.Background(), 1, model.ContentTypeFile, io.LimitReader(strings.NewReader("test"), 4), "test.pdf", 4, "application/pdf")
	require.NoError(t, err)
	assert.NotEmpty(t, resp.URL)
	assert.Equal(t, model.ContentTypeFile, resp.ContentType)
	assert.Equal(t, "test.pdf", resp.FileName)
	assert.Equal(t, "application/pdf", resp.FileType)

	oss.AssertExpectations(t)
}

func TestUploadService_ImageSuccess(t *testing.T) {
	oss := new(MockOSS)
	oss.On("PutObject", mock.Anything, mock.AnythingOfType("string"), mock.Anything, "image/webp", int64(10)).
		Return("https://bucket.oss-cn-hangzhou.aliyuncs.com/uploads/image/1/photo.webp", nil)

	svc := newTestUploadService(oss)

	resp, err := svc.Upload(context.Background(), 1, model.ContentTypeImage, io.LimitReader(strings.NewReader("fake-image"), 10), "photo.webp", 10, "image/webp")
	require.NoError(t, err)
	assert.Equal(t, model.ContentTypeImage, resp.ContentType)

	oss.AssertExpectations(t)
}

func TestUploadService_OSSFailure(t *testing.T) {
	oss := new(MockOSS)
	oss.On("PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", errors.New("OSS unavailable"))

	svc := newTestUploadService(oss)

	_, err := svc.Upload(context.Background(), 1, model.ContentTypeFile, nil, "doc.pdf", 100, "application/pdf")
	assert.Error(t, err)

	oss.AssertExpectations(t)
}
