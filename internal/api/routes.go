package api

import (
	"contrib.rocks/internal/api/image"
	"contrib.rocks/internal/service"
	"github.com/gin-gonic/gin"
)

type Router interface {
	gin.IRouter
}

func Setup(r Router, sp *service.ServicePack) {
	imageApi := image.New(sp.ContributorsService, sp.ImageService, sp.UsageService)
	r.GET("/image", imageApi.Get)
}
