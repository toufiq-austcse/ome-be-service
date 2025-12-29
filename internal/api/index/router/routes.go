package router

import (
	"github.com/gin-gonic/gin"
	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/index/controller"
)

func Setup(group *gin.RouterGroup, omeController *controller.OmeController) {
	group.GET("", controller.Index())
	group.POST("/ome/webhook", omeController.Webhook)
	group.POST("/ome/create", omeController.CreateStream)
	group.POST("/ome/startPush", omeController.StartPush)
	group.POST("/ome/stopPush", omeController.StopPush)

}
