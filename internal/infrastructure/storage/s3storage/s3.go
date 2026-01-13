package s3storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types" // ✅ enum de ACL
	"github.com/aws/smithy-go"
)

type Options struct {
	Bucket         string
	Region         string
	Endpoint       string // opcional (MinIO/Wasabi etc.)
	BaseURL        string // opcional — para links públicos/HTTP
	ForcePathStyle bool   // true para MinIO
	ACL            string // ex.: "private" (default), "public-read"
}

type Storage struct {
	s3     *s3.Client
	upl    *manager.Uploader
	dl     *manager.Downloader
	bucket string
	base   string
	acl    string
}

func New(ctx context.Context, opts Options, cfgFns ...func(*config.LoadOptions) error) (*Storage, error) {
	if strings.TrimSpace(opts.Bucket) == "" {
		return nil, errors.New("s3storage: bucket obrigatório")
	}
	if strings.TrimSpace(opts.Region) == "" {
		return nil, errors.New("s3storage: region obrigatória")
	}

	// Config AWS
	loaders := []func(*config.LoadOptions) error{
		config.WithRegion(opts.Region),
	}
	loaders = append(loaders, cfgFns...)

	// Endpoint custom + path style (MinIO/Wasabi/etc.)
	if opts.Endpoint != "" {
		loaders = append(loaders, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               strings.TrimRight(opts.Endpoint, "/"),
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		))
	}

	awscfg, err := config.LoadDefaultConfig(ctx, loaders...)
	if err != nil {
		return nil, fmt.Errorf("s3storage: falha ao carregar config AWS: %w", err)
	}

	client := s3.NewFromConfig(awscfg, func(o *s3.Options) {
		if opts.ForcePathStyle {
			o.UsePathStyle = true
		}
	})

	return &Storage{
		s3:     client,
		upl:    manager.NewUploader(client),
		dl:     manager.NewDownloader(client),
		bucket: opts.Bucket,
		base:   strings.TrimRight(opts.BaseURL, "/"),
		acl:    defaultACL(opts.ACL),
	}, nil
}

func defaultACL(v string) string {
	if v == "" {
		return "private"
	}
	return v
}

func (s *Storage) Save(ctx context.Context, key string, r io.Reader) (string, error) {
	key = sanitizeKey(key)
	if key == "" {
		return "", errors.New("s3storage: key vazio")
	}

	_, err := s.upl.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   r,
		ACL:    s3types.ObjectCannedACL(s.acl), // ✅ enum correto no SDK v2
	})
	if err != nil {
		return "", fmt.Errorf("s3storage: upload: %w", err)
	}

	return s.URL(key), nil
}

func (s *Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	key = sanitizeKey(key)
	if key == "" {
		return nil, errors.New("s3storage: key vazio")
	}

	// ✅ Simplificado: GetObject devolve um Body (io.ReadCloser)
	resp, err := s.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3storage: download: %w", err)
	}
	return resp.Body, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	key = sanitizeKey(key)
	if key == "" {
		return errors.New("s3storage: key vazio")
	}
	_, err := s.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3storage: delete: %w", err)
	}

	// ✅ Esperar remoção consistente (Waiter do SDK v2)
	waiter := s3.NewObjectNotExistsWaiter(s.s3)
	// timeout total de espera (ajuste se preferir)
	if werr := waiter.Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, 10*time.Second); werr != nil {
		// Não é fatal para a deleção em si, mas informamos
		return fmt.Errorf("s3storage: waiter not-exists: %w", werr)
	}
	return nil
}

func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	key = sanitizeKey(key)
	if key == "" {
		return false, errors.New("s3storage: key vazio")
	}
	_, err := s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}
	var apiErr *smithy.GenericAPIError
	if errors.As(err, &apiErr) && (apiErr.Code == "NotFound" || apiErr.Code == "404") {
		return false, nil
	}
	// alguns provedores retornam 404 encapsulado na mensagem
	if strings.Contains(strings.ToLower(err.Error()), "notfound") || strings.Contains(err.Error(), "404") {
		return false, nil
	}
	return false, fmt.Errorf("s3storage: head: %w", err)
}

func (s *Storage) URL(key string) string {
	key = sanitizeKey(key)
	if s.base == "" {
		// Sem base pública: devolve s3://bucket/key
		return fmt.Sprintf("s3://%s/%s", s.bucket, key)
	}
	u, _ := url.Parse(s.base)
	u.Path = path.Join(u.Path, key)
	return u.String()
}

func sanitizeKey(k string) string {
	k = strings.TrimSpace(k)
	k = strings.TrimLeft(k, "/")
	k = strings.ReplaceAll(k, "\\", "/")
	k = path.Clean(k)
	if k == "." || k == ".." || strings.HasPrefix(k, "../") {
		return ""
	}
	return k
}
