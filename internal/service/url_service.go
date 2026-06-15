package service

import (
	"context"
	"go-url-shortener/pkg/obfuscator"
)

type Repository interface {
	Save(ctx context.Context, longURL string) (uint64, error)
	GetLongURL(ctx context.Context, token string, id uint64) (string, error)
}

type URLService struct {
	repo Repository
}

func NewURLService(repo Repository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Shorten(ctx context.Context, longURL string) (string, error) {
	id, err := s.repo.Save(ctx, longURL)
	if err != nil {
		return "", err
	}
	return obfuscator.Encode(id), nil
}

func (s *URLService) Resolve(ctx context.Context, token string) (string, error) {
	id, err := obfuscator.Decode(token)
	if err != nil {
		return "", err
	}
	return s.repo.GetLongURL(ctx, token, id)
}
