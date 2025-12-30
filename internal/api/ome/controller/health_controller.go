package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/toufiq-austcse/go-api-boilerplate/config"
	"net/http"
)

type OmeWebhookRequest struct {
	Client struct {
		Address   string `json:"address"`
		Port      int    `json:"port"`
		RealIP    string `json:"real_ip"`
		UserAgent string `json:"user_agent"`
	} `json:"client"`
	Request struct {
		Direction string `json:"direction"`
		NewUrl    string `json:"new_url"`
		Protocol  string `json:"protocol"`
		Status    string `json:"status"` // "opening", "closing"
		Time      string `json:"time"`
		Url       string `json:"url"` // We extract name from here

	} `json:"request"`
}

// Index hosts godoc
// @Summary  Health Check
// @Tags     Index
// @Accept   json
// @Produce  json
// @Success  200
// @Router   / [get]
func Index() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": config.AppConfig.APP_NAME + " is Running",
		})
	}
}

func Webhook() gin.HandlerFunc {
	return func(context *gin.Context) {
		// Print headers
		fmt.Println("=== Webhook Request Headers ===")
		for key, values := range context.Request.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}

		var payload OmeWebhookRequest

		err := context.BindJSON(&payload)
		if err != nil {
			fmt.Println("Error parsing JSON payload:", err)
			return
		}

		// Print the parsed payload
		fmt.Println("=== Webhook Request Body (Parsed) ===")
		fmt.Printf("%+v\n", payload)
		fmt.Println("==============================")

		context.JSON(http.StatusOK, gin.H{
			"allowed": true,
		})
	}

}
