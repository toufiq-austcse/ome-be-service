package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/toufiq-austcse/go-api-boilerplate/config"
	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/index/model"
	"github.com/toufiq-austcse/go-api-boilerplate/internal/api/index/service"
	"github.com/toufiq-austcse/go-api-boilerplate/pkg/api_response"
	"github.com/toufiq-austcse/go-api-boilerplate/pkg/http_clients"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StartPushRequest struct {
	StreamID string `json:"stream_id" binding:"required"`
	RtmpUrl  string `json:"rtmp_url" binding:"required"`
}
type StopPushRequest struct {
	StreamID string `json:"stream_id" binding:"required"`
}
type CreateStreamRequest struct {
	ExternalId string `json:"external_id" binding:"required"`
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

		err := controller.omeHttpClient.DeleteStream(streamName, c.Request.Header.Get("X-Forwarded-For"))
		if err != nil {
			fmt.Println("Error deleting existing stream:", err)

			c.JSON(http.StatusOK, gin.H{
				"allowed": false,
			})

		}

		_, err = controller.omeService.UpdateStreamByID(c, objId, bson.M{
			"status":            payload.Request.Status,
			"protocol":          payload.Request.Protocol,
			"server_ip_address": c.Request.Header.Get("X-Forwarded-For"),
		})
		if err != nil {
			return
		}
	} else {
		_, err = controller.omeService.UpdateStreamByID(c, objId, bson.M{
			"status": payload.Request.Status,
		})
		if err != nil {
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"allowed": true,
	})
}

func (controller *OmeController) CreateStream(c *gin.Context) {
	var body CreateStreamRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Error binding JSON:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	var resultStream *model.Stream

	existingStream, err := controller.omeService.GetStreamByExternalId(c, body.ExternalId)
	if err != nil {
		fmt.Println("Error checking existing stream:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to check existing stream",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	if existingStream != nil {
		resultStream = existingStream
	} else {
		stream, err := controller.omeService.CreateStream(c, &model.Stream{
			Status:     "initiated",
			ExternalId: body.ExternalId,
		})
		if err != nil {
			fmt.Println("Error creating stream:", err)
			errResponse := api_response.BuildErrorResponse(
				http.StatusInternalServerError,
				"Failed to create stream",
				err.Error(), "")
			c.JSON(errResponse.Code, errResponse)
			return
		}
		resultStream = stream
	}

	whipUrl := fmt.Sprintf("%s/app/%s?direction=whip", config.AppConfig.OME_SERVER_BASE_URL, resultStream.Id.Hex())
	apiResponseBody := api_response.BuildResponse(http.StatusOK, "Stream created successfully", map[string]interface{}{
		"_id":         resultStream.Id.Hex(),
		"status":      resultStream.Status,
		"external_id": resultStream.ExternalId,
		"whip_url":    whipUrl,
		"created_at":  resultStream.CreatedAt,
		"updated_at":  resultStream.UpdatedAt,
	})
	c.JSON(apiResponseBody.Code, apiResponseBody)

}

func (controller *OmeController) StopStream(c *gin.Context) {
	//streamId := c.Param("id")
	//objId, err := primitive.ObjectIDFromHex(streamId)
	//if err != nil {
	//	fmt.Println("Error getting stream by name:", err)
	//	return
	//}
	//
	//_, err = controller.omeService.FindByID(c, objId)
	//if err != nil {
	//	fmt.Println("Error finding stream by ID:", err)
	//	return
	//}
	//
	//c.JSON(http.StatusOK, gin.H{
	//	"message": "Stream stopped successfully",
	//})

}

func (controller *OmeController) StartPush(c *gin.Context) {
	var body StartPushRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Error binding JSON:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	objId, err := primitive.ObjectIDFromHex(body.StreamID)
	if err != nil {
		fmt.Println("Error getting stream by ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid stream ID",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	existingStream, err := controller.omeService.FindStreamById(c, objId)
	if err != nil {
		fmt.Println("Error finding stream by ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Stream not found",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	if existingStream == nil {
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Stream not found",
			"Stream does not exist", "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	existingActivePush, err := controller.omeService.FindPushByStreamIdAndStatus(c, objId, "active")
	if err != nil {
		fmt.Println("Error finding stream by ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Stream not found",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	if existingActivePush != nil {
		err := controller.omeHttpClient.StopPush(existingStream.ServerIpAddress, existingActivePush.Id.Hex())
		if err != nil {
			fmt.Println("Error stopping existing push:", err)
			errResponse := api_response.BuildErrorResponse(
				http.StatusInternalServerError,
				"Failed to stop existing push",
				err.Error(), "")
			c.JSON(errResponse.Code, errResponse)
			return
		}
		_, err = controller.omeService.UpdatePushByID(
			c,
			existingActivePush.Id, bson.M{
				"status": "inactive",
			})
		if err != nil {
			fmt.Println("Error updating existing push status:", err)
			errResponse := api_response.BuildErrorResponse(
				http.StatusInternalServerError,
				"Failed to update existing push status",
				err.Error(), "")
			c.JSON(errResponse.Code, errResponse)
			return
		}
	}
	pushId := primitive.NewObjectID()

	_, err = controller.omeHttpClient.StartPush(
		existingStream.ServerIpAddress,
		existingStream.Id.Hex(),
		body.RtmpUrl, pushId.Hex())
	if err != nil {
		fmt.Println("Error starting push:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to start push",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	_, err = controller.omeService.CreatePush(c, &model.Push{
		StreamId: objId,
		RtmpUrl:  body.RtmpUrl,
		Status:   "active",
		Id:       pushId,
	})
	if err != nil {
		fmt.Println("Error creating push record:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to create push record",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Push started successfully",
		"push_id":   pushId.Hex(),
		"rtmp_url":  body.RtmpUrl,
		"server_ip": existingStream.ServerIpAddress,
		"stream_id": body.StreamID,
	})

}

func (controller *OmeController) StopPush(c *gin.Context) {
	var body StopPushRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		fmt.Println("Error binding JSON:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	objId, err := primitive.ObjectIDFromHex(body.StreamID)
	if err != nil {
		fmt.Println("Error getting stream by ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusBadRequest,
			"Invalid stream ID",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	existingStream, err := controller.omeService.FindStreamById(c, objId)
	if err != nil {
		fmt.Println("Error finding stream by ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Stream not found",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	if existingStream == nil {
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Stream not found",
			"Stream does not exist", "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	existingActivePush, err := controller.omeService.FindPushByStreamIdAndStatus(c, objId, "active")
	if err != nil {
		fmt.Println("Error finding active push by stream ID:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Active push not found",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	if existingActivePush == nil {
		errResponse := api_response.BuildErrorResponse(
			http.StatusNotFound,
			"Active push not found",
			"No active push for this stream", "")
		c.JSON(errResponse.Code, errResponse)
		return
	}

	err = controller.omeHttpClient.StopPush(existingStream.ServerIpAddress, existingActivePush.Id.Hex())
	if err != nil {
		fmt.Println("Error stopping push:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to stop push",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	_, err = controller.omeService.UpdatePushByID(
		c,
		existingActivePush.Id, bson.M{
			"status": "inactive",
		})
	if err != nil {
		fmt.Println("Error updating push status:", err)
		errResponse := api_response.BuildErrorResponse(
			http.StatusInternalServerError,
			"Failed to update push status",
			err.Error(), "")
		c.JSON(errResponse.Code, errResponse)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Push stopped successfully",
	})

}
