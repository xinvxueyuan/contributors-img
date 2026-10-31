package app

import (
	"encoding/json"
	"fmt"

	"contrib.rocks/internal/api"
	"contrib.rocks/internal/apierror"
	"contrib.rocks/internal/config"
	"contrib.rocks/internal/logger"
	"contrib.rocks/internal/service"
	"contrib.rocks/internal/tracing"
	"contrib.rocks/pkg/compress"
	"contrib.rocks/pkg/env"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StartServer() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %s", err.Error())
	}
	fmt.Printf("config: %+v\n", cfg)

	if cfg.Env == env.EnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}

	tracing.Init()

	sp, err := service.NewServicePack(cfg)
	if err != nil {
		return fmt.Errorf("failed to create service pack: %s", err.Error())
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(config.Middleware(cfg))
	r.Use(tracing.Middleware(cfg))
	r.Use(logger.Middleware(cfg))
	r.Use(apierror.Middleware())
	r.Use(requestLogger())
	r.Use(compress.Compress())

	api.Setup(r, sp)

	fmt.Printf("Listening on http://localhost:%s\n", cfg.Port)
	return r.Run(fmt.Sprintf(":%s", cfg.Port))
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.LoggerFromContext(c.Request.Context()).Debug("request.start",
			zap.String("method", c.Request.Method),
			zap.String("host", c.Request.Host),
			zap.String("url", c.Request.URL.String()),
			zap.String("userAgent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
		)
		// log all headers
		headers, _ := json.Marshal(c.Request.Header)
		logger.LoggerFromContext(c.Request.Context()).Debug("request.headers", zap.String("headers", string(headers)))
		c.Next()
		logger.LoggerFromContext(c.Request.Context()).Debug("request.end",
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}
