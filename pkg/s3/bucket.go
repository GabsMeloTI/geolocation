package bucket

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var S3Client *s3.S3

func InitS3Client() {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if accessKeyID == "" || secretAccessKey == "" || region == "" {
		panic("AWS credentials or region are not set in environment variables")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create session: %v", err))
	}

	S3Client = s3.New(sess)
}

func UploadFileToS3(fileBytes []byte, fileName, bucket, contentType string) (string, error) {
	InitS3Client()
	if S3Client == nil {
		return "", fmt.Errorf("S3 client is not initialized")
	}

	reader := bytes.NewReader(fileBytes)

	_, err := S3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(fileName),
		Body:          reader,
		ContentLength: aws.Int64(int64(len(fileBytes))),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %v", err)
	}

	imageURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, fileName)
	return imageURL, nil
}

func DeleteFile(ctx context.Context, bucketName, key string) error {
	InitS3Client()
	if S3Client == nil {
		return fmt.Errorf("S3 client is not initialized")
	}

	_, err := S3Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %v", err)
	}

	err = S3Client.WaitUntilObjectNotExistsWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error waiting for file deletion: %v", err)
	}

	return nil
}
