package router

import (
	"github.com/gin-gonic/gin"
	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/index/controller"
)

func Setup(group *gin.RouterGroup, omeController *controller.OmeController) {
	group.GET("", controller.Index())
	group.POST("/ome/webhook", omeController.Webhook)
	group.POST("/ome/create", omeController.CreateStream)
	group.DELETE("/ome/:id", omeController.StopStream)
	group.POST("/ome/:id/startPush", omeController.StartPush)
	group.POST("/ome/:id/stopPush", omeController.StopPush)

}
