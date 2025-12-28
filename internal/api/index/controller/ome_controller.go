package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/toufiq-austcse/go-api-boilerplate/config"
	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/index/service"
	"github.com/toufiq-austcse/go-api-boilerplate/pkg/api_response"
	"github.com/toufiq-austcse/go-api-boilerplate/pkg/http_clients"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

type StartPushRequest struct {
	RtmpUrl string `json:"rtmp_url" binding:"required"`
}
type OmeController struct {
	omeService    *service.OmeService
	omeHttpClient *http_clients.OmeHTTPClient
}

func NewOmeController(omeService *service.OmeService, omeHttpClient *http_clients.OmeHTTPClient) *OmeController {
	return &OmeController{
		omeService:    omeService,
		omeHttpClient: omeHttpClient,
	}
}

func (controller *OmeController) Webhook(c *gin.Context) {
	// Print headers
	fmt.Println("=== Webhook Request Headers ===")
	for key, values := range c.Request.Header {
		for _, value := range values {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	var payload OmeWebhookRequest

	err := c.BindJSON(&payload)
	if err != nil {
		fmt.Println("Error parsing JSON payload:", err)
		return
	}

	// Print the parsed payload
	fmt.Println("=== Webhook Request Body (Parsed) ===")
	fmt.Printf("%+v\n", payload)
	fmt.Println("==============================")

	streamName := controller.omeService.GetStreamName(payload.Request.Url)
	objId, err := primitive.ObjectIDFromHex(streamName)
	if err != nil {
		fmt.Println("Error getting stream by name:", err)
		return
	}
	if payload.Request.Status == "opening" {
		err = controller.omeService.UpdateStreamById(c, objId, map[string]interface{}{
			"status":     payload.Request.Status,
			"ip_address": c.Request.Header.Get("X-Forwarded-For"),
			"protocol":   payload.Request.Protocol,
		})
		if err != nil {
			return
		}
	} else {
		err = controller.omeService.UpdateStatusByName(c, streamName, payload.Request.Status)
		if err != nil {
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"allowed": true,
	})
}

func (controller *OmeController) CreateStream(c *gin.Context) {
	newStream, err := controller.omeService.Create(c)
	if err != nil {
		return
	}

	// Convert ObjectID to hex string
	var streamId string
	if objID, ok := newStream.(map[string]interface{})["_id"].(primitive.ObjectID); ok {
		streamId = objID.Hex()
	} else {
		streamId = fmt.Sprintf("%v", newStream.(map[string]interface{})["_id"])
	}

	whipUrl := fmt.Sprintf("%s/app/%s?direction=whip", config.AppConfig.OME_SERVER_BASE_URL, streamId)

	newStreamRes := api_response.BuildResponse(
		http.StatusCreated,
		http.StatusText(http.StatusCreated),
		map[string]interface{}{
			"_id":      streamId,
			"status":   newStream.(map[string]interface{})["status"],
			"whip_url": whipUrl,
		},
	)

	c.JSON(newStreamRes.Code, newStreamRes)
}

func (controller *OmeController) StopStream(c *gin.Context) {
	streamId := c.Param("id")
	objId, err := primitive.ObjectIDFromHex(streamId)
	if err != nil {
		fmt.Println("Error getting stream by name:", err)
		return
	}

	_, err = controller.omeService.FindByID(c, objId)
	if err != nil {
		fmt.Println("Error finding stream by ID:", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stream stopped successfully",
	})

}

func (controller *OmeController) StartPush(c *gin.Context) {
	streamId := c.Param("id")

	objId, err := primitive.ObjectIDFromHex(streamId)
	if err != nil {
		fmt.Println("Error getting stream by name:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid stream ID",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
	}
	body := StartPushRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Error binding JSON:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	stream, err := controller.omeService.FindByID(c, objId)
	if err != nil {
		fmt.Println("Error finding stream by ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Stream not found",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	response, err := controller.omeHttpClient.StartPush(
		stream["ip_address"].(string),
		streamId+"_rtmp", body.RtmpUrl)
	if err != nil {
		fmt.Println("Error starting push:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to start push",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	err = controller.omeService.UpdateStreamById(c, objId, map[string]interface{}{
		"push_id": response.Response.ID,
	})
	if err != nil {
		fmt.Println("Error updating stream with push ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to update stream with push ID",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

}

func (controller *OmeController) StopPush(c *gin.Context) {
	streamId := c.Param("id")

	objId, err := primitive.ObjectIDFromHex(streamId)
	if err != nil {
		fmt.Println("Error getting stream by name:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid stream ID",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	stream, err := controller.omeService.FindByID(c, objId)
	if err != nil {
		fmt.Println("Error finding stream by ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Stream not found",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	err = controller.omeHttpClient.StopPush(
		stream["ip_address"].(string),
		stream["push_id"].(string))
	if err != nil {
		fmt.Println("Error stopping push:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to stop push",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Push stopped successfully",
	})

}
