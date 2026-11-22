package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/google/uuid"
)

// ErrInvalidContent means the photo type is not jpeg, png, or webp.
var ErrInvalidContent = errors.New("invalid image")

// Config is the Cloudflare R2 bucket the piece photos live in.
type Config struct {
	Bucket          string `yaml:"bucket" env:"R2_BUCKET"`
	AccountID       string `yaml:"accountId" env:"R2_ACCOUNT_ID"`
	AccessKeyID     string `yaml:"accessKeyId" env:"R2_ACCESS_KEY_ID"`
	AccessKeySecret string `yaml:"accessKeySecret" env:"R2_ACCESS_KEY_SECRET"`
	Env             string `yaml:"env" env:"R2_ENV" env-default:"local"`
}

type r2Store struct {
	client *s3.Client
	cfg    Config
}

type unavailableStore struct{}

func New(cfg Config) Store {
	if cfg.Bucket == "" || cfg.AccountID == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" {
		return unavailableStore{}
	}
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.AccessKeySecret, "")),
	)
	if err != nil {
		return unavailableStore{}
	}
	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID))
		options.UsePathStyle = true
	})
	return &r2Store{client: client, cfg: cfg}
}

func (s *r2Store) PresignUpload(ctx context.Context, googleID, contentType string) (Upload, error) {
	extension, ok := extensionFor(contentType)
	if !ok || !safeID(googleID) {
		return Upload{}, ErrInvalidContent
	}
	key := fmt.Sprintf("items/%s/%s%s", googleID, uuid.NewString(), extension)
	presigner := s3.NewPresignClient(s.client)
	signed, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return Upload{}, err
	}
	readURL, err := s.PresignRead(ctx, key)
	if err != nil {
		return Upload{}, err
	}
	return Upload{
		URL:         signed.URL,
		Key:         key,
		ReadURL:     readURL,
		ContentType: contentType,
	}, nil
}

func (s *r2Store) PresignRead(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", nil
	}
	presigner := s3.NewPresignClient(s.client)
	signed, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return "", err
	}
	return signed.URL, nil
}

func (s *r2Store) BelongsTo(googleID, key string) bool {
	return belongs(googleID, key)
}

func (s unavailableStore) PresignUpload(context.Context, string, string) (Upload, error) {
	return Upload{}, ErrNotConfigured
}

func (unavailableStore) PresignRead(context.Context, string) (string, error) {
	return "", nil
}

func (unavailableStore) BelongsTo(googleID, key string) bool {
	return belongs(googleID, key)
}

func extensionFor(contentType string) (string, bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}

func safeID(id string) bool {
	if id == "" || len(id) > 80 || strings.Contains(id, "/") {
		return false
	}
	return true
}
