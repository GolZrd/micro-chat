package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage interface {
	Upload(ctx context.Context, filename string, reader io.Reader, size int64, contentType string) (string, error)
}

type minioStorage struct {
	client   *minio.Client
	bucket   string
	endpoint string
}

func NewMinioStorage(endpoint, accessKey, secretKey, bucket string) (Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false, // Пока не используем SSL
	})

	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}

		// Публичный доступ на чтение
		policy := fmt.Sprintf(`{
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Principal": {"AWS": ["*"]},
                "Action": ["s3:GetObject"],
                "Resource": ["arn:aws:s3:::%s/*"]
            }]
        }`, bucket)

		err = client.SetBucketPolicy(ctx, bucket, policy)
		if err != nil {
			return nil, fmt.Errorf("set bucket policy: %w", err)
		}
	}

	return &minioStorage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}, nil
}

func (s *minioStorage) Upload(ctx context.Context, filename string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, filename, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("upload file: %w", err)
	}

	url := fmt.Sprintf("/files/%s/%s", s.bucket, filename)

	return url, nil
}
