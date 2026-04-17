package storage

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"gin-scaffold/config"
)

var s3MetaKeyRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

func isS3NotFound(err error) bool {
	var re *awshttp.ResponseError
	if !errors.As(err, &re) {
		return false
	}
	if re.Response == nil {
		return false
	}
	return re.Response.StatusCode == http.StatusNotFound
}

// S3Provider S3 兼容对象存储（含 MinIO）。
type S3Provider struct {
	bucket        string
	client        *s3.Client
	presignClient *s3.PresignClient
	secret        []byte
}

func NewS3Provider(cfg *config.StorageConfig) (*S3Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage s3: nil config")
	}
	if strings.TrimSpace(cfg.SignSecret) == "" {
		return nil, fmt.Errorf("storage s3: sign secret is empty")
	}
	region := strings.TrimSpace(cfg.S3Region)
	if region == "" {
		region = "us-east-1"
	}
	endpoint := strings.TrimSpace(cfg.S3Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("storage s3: endpoint is empty")
	}
	bucket := strings.TrimSpace(cfg.S3Bucket)
	if bucket == "" {
		return nil, fmt.Errorf("storage s3: bucket is empty")
	}
	accessKey := strings.TrimSpace(cfg.S3AccessKey)
	secretKey := strings.TrimSpace(cfg.S3SecretKey)
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("storage s3: access key or secret key is empty")
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}
	if cfg.S3Insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} //nolint:gosec
		opts = append(opts, awsconfig.WithHTTPClient(&http.Client{Transport: tr, Timeout: 30 * time.Second}))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = cfg.S3PathStyle
	})

	return &S3Provider{
		bucket:        bucket,
		client:        client,
		presignClient: s3.NewPresignClient(client),
		secret:        []byte(cfg.SignSecret),
	}, nil
}

func (p *S3Provider) PresignPutURL(ctx context.Context, key string, contentType string, expire time.Duration, opts *PresignPutOptions) (string, string, map[string]string, error) {
	if p == nil || p.presignClient == nil {
		return "", "", nil, fmt.Errorf("storage s3: presign client not initialized")
	}
	if expire <= 0 {
		expire = 15 * time.Minute
	}
	k := normalizeKey(key)
	if k == "" {
		return "", "", nil, fmt.Errorf("storage s3: empty key")
	}
	ct := strings.TrimSpace(strings.Split(contentType, ";")[0])
	if ct == "" {
		return "", "", nil, fmt.Errorf("storage s3: content type is empty")
	}
	in := &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(k),
		ContentType: aws.String(ct),
	}
	if opts != nil {
		if opts.ContentLength > 0 {
			in.ContentLength = aws.Int64(opts.ContentLength)
		}
		if len(opts.Metadata) > 0 {
			meta := make(map[string]string, len(opts.Metadata))
			for mk, mv := range opts.Metadata {
				mk = strings.ToLower(strings.TrimSpace(mk))
				if mk == "" || !s3MetaKeyRe.MatchString(mk) {
					return "", "", nil, fmt.Errorf("storage s3: invalid metadata key %q", mk)
				}
				mv = strings.TrimSpace(mv)
				if len(mv) > 2048 {
					return "", "", nil, fmt.Errorf("storage s3: metadata value too long for key %q", mk)
				}
				meta[mk] = mv
			}
			in.Metadata = meta
		}
	}
	out, err := p.presignClient.PresignPutObject(ctx, in, s3.WithPresignExpires(expire))
	if err != nil {
		return "", "", nil, err
	}
	headers := make(map[string]string, len(out.SignedHeader))
	for hk, vv := range out.SignedHeader {
		if len(vv) == 0 {
			continue
		}
		headers[hk] = vv[0]
	}
	return out.Method, out.URL, headers, nil
}

func (p *S3Provider) StatObject(ctx context.Context, key string) (*ObjectStat, error) {
	if p == nil || p.client == nil {
		return nil, fmt.Errorf("storage s3: client not initialized")
	}
	k := normalizeKey(key)
	if k == "" {
		return nil, fmt.Errorf("storage s3: empty key")
	}
	out, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(k),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, ErrObjectNotExist
		}
		return nil, err
	}
	meta := make(map[string]string, len(out.Metadata))
	for mk, mv := range out.Metadata {
		meta[strings.ToLower(strings.TrimSpace(mk))] = strings.TrimSpace(mv)
	}
	return &ObjectStat{
		Size:         aws.ToInt64(out.ContentLength),
		ContentType:  aws.ToString(out.ContentType),
		Metadata:     meta,
		ETag:         strings.Trim(aws.ToString(out.ETag), `"`),
		DeleteMarker: aws.ToBool(out.DeleteMarker),
	}, nil
}

func (p *S3Provider) Ready(ctx context.Context) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("storage s3: client not initialized")
	}
	_, err := p.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(p.bucket)})
	return err
}

func (p *S3Provider) Put(ctx context.Context, key string, reader io.Reader) error {
	return p.PutContentType(ctx, key, "", reader)
}

func (p *S3Provider) PutContentType(ctx context.Context, key string, contentType string, reader io.Reader) error {
	k := normalizeKey(key)
	if k == "" {
		return fmt.Errorf("storage s3: empty key")
	}
	in := &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(k),
		Body:   reader,
	}
	if strings.TrimSpace(contentType) != "" {
		in.ContentType = aws.String(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	}
	_, err := p.client.PutObject(ctx, in)
	return err
}

func (p *S3Provider) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	k := normalizeKey(key)
	if k == "" {
		return nil, fmt.Errorf("storage s3: empty key")
	}
	out, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(k),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (p *S3Provider) Delete(ctx context.Context, key string) error {
	k := normalizeKey(key)
	if k == "" {
		return fmt.Errorf("storage s3: empty key")
	}
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(k),
	})
	return err
}

func (p *S3Provider) Sign(key string, expireSec int64) (string, error) {
	return signDownload(p.secret, key, expireSec)
}

func (p *S3Provider) Verify(key string, expires int64, sig string) bool {
	return verifyDownload(p.secret, key, expires, sig)
}
