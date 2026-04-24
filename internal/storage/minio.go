package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client     *minio.Client
	bucketName string
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) *MinioClient {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Printf("Failed to initialize MinIO client: %v", err)
		return nil
	}

	bucketName := "avatar"
	ctx := context.Background()
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("Bucket %s already exists\n", bucketName)
		} else {
			log.Printf("Error checking/creating bucket: %v\n", err)
		}
	} else {
		log.Printf("Successfully created bucket %s\n", bucketName)
	}

	// Set public read policy
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": ["s3:GetObject"],
				"Effect": "Allow",
				"Principal": "*",
				"Resource": ["arn:aws:s3:::%s/*"],
				"Sid": ""
			}
		]
	}`, bucketName)
	err = minioClient.SetBucketPolicy(ctx, bucketName, policy)
	if err != nil {
		log.Printf("Error setting bucket policy: %v\n", err)
	}

	return &MinioClient{
		client:     minioClient,
		bucketName: bucketName,
	}
}

func (m *MinioClient) UploadAvatar(ctx context.Context, file *multipart.FileHeader) (string, error) {
	if m == nil {
		return "", fmt.Errorf("minio client not initialized")
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename using timestamp to avoid caching issues and conflicts
	userIDVal := ctx.Value("user_id")
	var fileName string
	if userIDVal != nil {
		fileName = fmt.Sprintf("avatar-%v-%d-%s", userIDVal, time.Now().Unix(), file.Filename)
	} else {
		fileName = fmt.Sprintf("avatar-%d-%s", time.Now().Unix(), file.Filename)
	}

	info, err := m.client.PutObject(ctx, m.bucketName, fileName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", err
	}

	// We return the path that should be prefixed with the MinIO base URL or gateway
	return fmt.Sprintf("/%s/%s", m.bucketName, info.Key), nil
}
