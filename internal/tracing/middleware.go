package tracing

import (
	"contrib.rocks/internal/config"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

func Middleware(cfg *config.Config) gin.HandlerFunc {

	return func(c *gin.Context) {
		originalCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(originalCtx)
		}()
		ctx := otel.GetTextMapPropagator().Extract(originalCtx, propagation.HeaderCarrier(c.Request.Header))
		ctx, span := Tracer().Start(ctx, "api.http")
		defer span.End()
		span.SetAttributes(attribute.String("/app/environment", string(cfg.Env)))

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
