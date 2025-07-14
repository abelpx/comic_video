package minio

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioClient MinIO 客户端接口
type MinioClient interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error)
	Delete(ctx context.Context, objectName string) error
	GetURL(objectName string) string
	PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (*url.URL, error)
	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, contentType string) error
	DeleteObject(ctx context.Context, bucketName, objectName string) error
}

var _ MinioClient = (*client)(nil)

type client struct {
	client     *minio.Client
	bucketName string
	publicHost string // 用于拼接外部可访问URL
}

// NewClient 初始化 MinIO 客户端
func NewClient(endpoint, accessKey, secretKey, bucket, publicHost string, useSSL bool) MinioClient {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	exists, err := cli.BucketExists(ctx, bucket)
	if err != nil {
		panic(err)
	}
	if !exists {
		err = cli.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			panic(err)
		}
	}

	return &client{
		client:     cli,
		bucketName: bucket,
		publicHost: publicHost,
	}
}

// Upload 上传文件
func (c *client) Upload(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) (string, error) {
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return c.GetURL(objectName), nil
}

// Delete 删除文件
func (c *client) Delete(ctx context.Context, objectName string) error {
	return c.client.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
}

// GetURL 获取文件外部可访问URL
func (c *client) GetURL(objectName string) string {
	if c.publicHost != "" {
		return c.publicHost + "/" + c.bucketName + "/" + objectName
	}
	return ""
}

// PresignedURL 获取带签名的临时访问URL
func (c *client) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedGetObject(ctx, c.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

// PresignedGetObject 获取带签名的临时访问URL（返回URL对象）
func (c *client) PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (*url.URL, error) {
	return c.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
}

// PutObject 上传对象
func (c *client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, contentType string) error {
	_, err := c.client.PutObject(ctx, bucketName, objectName, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// DeleteObject 删除对象
func (c *client) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	return c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
} 