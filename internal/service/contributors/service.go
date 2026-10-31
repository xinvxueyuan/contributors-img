package contributors

import (
	"context"
	"fmt"

	"contrib.rocks/internal/logger"
	"contrib.rocks/internal/service/internal/appcache"
	"contrib.rocks/internal/service/internal/cachekey"
	"contrib.rocks/internal/service/internal/cacheutil"
	"contrib.rocks/internal/tracing"
	"contrib.rocks/pkg/model"
	"github.com/google/go-github/v69/github"
)

type GitHubClientProvider interface {
	Get() *github.Client
}

type Service struct {
	ghProvider GitHubClientProvider
	cache      appcache.AppCache
}

func New(ghProvider GitHubClientProvider, cache appcache.AppCache) *Service {
	return &Service{ghProvider, cache}
}

func (s *Service) GetContributors(c context.Context, r *model.Repository) (*model.RepositoryContributors, error) {
	ctx, span := tracing.Tracer().Start(c, "contributors.Service.GetContributors")
	defer span.End()
	log := logger.LoggerFromContext(ctx)

	cacheKey := cachekey.ForContributors(r, "json")
	// restore cache
	var cache *model.RepositoryContributors
	err := s.cache.GetJSON(ctx, cacheKey, &cache)
	if err != nil {
		return nil, err
	}
	if cache != nil {
		log.Debug(fmt.Sprintf("restored contributors-json from cache: %s", cacheKey))
		return cache, nil
	}
	cacheutil.LogCacheMiss(ctx, "contributors-json", cacheKey)

	// get contributors from github
	gh := s.ghProvider.Get()
	data, err := fetchRepositoryContributors(gh, ctx, r)
	if err != nil {
		return nil, err
	}
	// save cache
	err = s.cache.SaveJSON(ctx, cacheKey, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
