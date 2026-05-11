package common

import (
	"context"
	"fmt"
	"io"

	osssdk "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	osscreds "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"

	"wetalk/config"
)

// Client OSS 客户端封装
type Client struct {
	client   *osssdk.Client
	bucket   string
	region   string
	endpoint string
}

// NewClient 根据 OSS 配置创建客户端
func NewClient(cfg config.OSSConfig) *Client {
	provider := osscreds.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret)

	ossCfg := osssdk.LoadDefaultConfig().
		WithCredentialsProvider(provider).
		WithRegion(cfg.Region)

	client := osssdk.NewClient(ossCfg)

	return &Client{
		client:   client,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		endpoint: cfg.Endpoint,
	}
}

// PutObject 上传对象到 OSS，返回可访问的 URL
func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, contentType string, contentLength int64) (string, error) {
	req := &osssdk.PutObjectRequest{
		Bucket:      osssdk.Ptr(c.bucket),
		Key:         osssdk.Ptr(key),
		Body:        body,
		ContentType: osssdk.Ptr(contentType),
		Acl:         osssdk.ObjectACLPublicRead,
	}

	if contentLength > 0 {
		req.ContentLength = osssdk.Ptr(contentLength)
	}

	_, err := c.client.PutObject(ctx, req)
	if err != nil {
		return "", fmt.Errorf("上传到 OSS 失败: %w", err)
	}

	return fmt.Sprintf("https://%s.%s/%s", c.bucket, c.endpoint, key), nil
}
