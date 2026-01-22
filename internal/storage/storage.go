package storage

import (
	"context"
	"io"
)

type ObjectInfo struct {
	ContentType string
	Size        int64
}

type Object struct {
	io.ReadCloser
	Info ObjectInfo
}

type Provider interface {
	EnsureBucket(ctx context.Context) error
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, key string) (*Object, error)
	Delete(ctx context.Context, key string) error
	CopyObject(ctx context.Context, srcKey, dstKey string) error
}
