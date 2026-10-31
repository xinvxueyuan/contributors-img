package service

import (
	"fmt"

	"contrib.rocks/internal/config"
	"contrib.rocks/internal/github"
	"contrib.rocks/internal/service/contributors"
	"contrib.rocks/internal/service/image"
	"contrib.rocks/internal/service/internal/appcache"
	"contrib.rocks/internal/service/usage"
)

type ServicePack struct {
	ContributorsService *contributors.Service
	UsageService        *usage.Service
	ImageService        *image.Service
}

func NewServicePack(cfg *config.Config) (*ServicePack, error) {
	ghProvider, err := newGitHubProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub provider: %w", err)
	}

	// This deployment has no object storage, so the cache lives in the process.
	cache := appcache.NewMemoryCache()

	return &ServicePack{
		ContributorsService: contributors.New(ghProvider, cache),
		ImageService:        image.New(cache),
		UsageService:        usage.New(),
	}, nil
}

// newGitHubProvider prefers GitHub App installation authentication when
// configured, falling back to a static token. config.Load already rejects a
// partial App configuration and the case where neither is set.
func newGitHubProvider(cfg *config.Config) (contributors.GitHubClientProvider, error) {
	if cfg.GitHubApp != nil {
		return github.NewProvider(cfg.GitHubApp.AppID, cfg.GitHubApp.InstallationID, cfg.GitHubApp.PrivateKey)
	}
	return github.NewTokenProvider(cfg.GitHubAuthToken), nil
}
