package http_clients

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type OmeHTTPClient struct {
	restyReq *resty.Request
}

type StartPushRequest struct {
	ID       string          `json:"id"`
	Stream   StartPushStream `json:"stream"`
	Protocol string          `json:"protocol"`
	URL      string          `json:"url"`
}

type StartPushStream struct {
	Name string `json:"name"`
}

type StartPushResponse struct {
	Message    string              `json:"message"`
	Response   PushResponseDetails `json:"response"`
	StatusCode int                 `json:"statusCode"`
}

type PushResponseDetails struct {
	App            string     `json:"app"`
	CreatedTime    string     `json:"createdTime"`
	FinishTime     string     `json:"finishTime"`
	ID             string     `json:"id"`
	IsConfig       bool       `json:"isConfig"`
	Protocol       string     `json:"protocol"`
	SentBytes      int64      `json:"sentBytes"`
	SentTime       int64      `json:"sentTime"`
	Sequence       int        `json:"sequence"`
	StartTime      string     `json:"startTime"`
	State          string     `json:"state"`
	Stream         StreamInfo `json:"stream"`
	TotalSentBytes int64      `json:"totalsentBytes"`
	TotalSentTime  int64      `json:"totalsentTime"`
	URL            string     `json:"url"`
	VHost          string     `json:"vhost"`
}

type StreamInfo struct {
	Name         string   `json:"name"`
	TrackIds     []string `json:"trackIds"`
	VariantNames []string `json:"variantNames"`
}

func NewOmeHTTPClient() *OmeHTTPClient {
	return &OmeHTTPClient{
		restyReq: resty.New().R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Authorization", "Basic b21lLWFjY2Vzcy10b2tlbg=="),
	}
}

func (c *OmeHTTPClient) StartPush(ip string, streamName string, rtmpUrl string, pushId string) (*StartPushResponse, error) {
	baseUrl := "http://" + ip + ":8081"

	// CreateStream request payload
	requestBody := StartPushRequest{
		ID: pushId,
		Stream: StartPushStream{
			Name: streamName + "_rtmp",
		},
		Protocol: "rtmp",
		URL:      rtmpUrl,
	}

	fmt.Println("baseUrl ", baseUrl)
	fmt.Println("requestBody", requestBody)

	var response StartPushResponse
	resp, err := c.restyReq.
		SetBody(requestBody).
		SetResult(&response).
		Post(baseUrl + "/v1/vhosts/default/apps/app:startPush")

	if err != nil {
		return nil, fmt.Errorf("failed to start push: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("start push failed with status code: %d", resp.StatusCode())
	}

	return &response, nil
}

func (c *OmeHTTPClient) GetBaseUrlFromIp(ip string) string {
	return "http://" + ip + ":8081"
}

func (c *OmeHTTPClient) StopPush(ip string, pushId string) error {
	baseUrl := c.GetBaseUrlFromIp(ip)

	fmt.Println("baseUrl ", baseUrl)
	fmt.Println("stopping push", pushId)
	resp, err := c.restyReq.
		SetBody(map[string]interface{}{
			"id": pushId,
		}).
		Post(baseUrl + "/v1/vhosts/default/apps/app:stopPush")

	if err != nil {
		return fmt.Errorf("failed to stop push: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("stop push failed with status code: %d", resp.StatusCode())
	}

	return nil
}
func (c *OmeHTTPClient) DeleteStream(streamName string, ip string) error {
	fmt.Println("Deleting stream:", streamName, "from IP:", ip)

	baseUrl := c.GetBaseUrlFromIp(ip)

	deleteResponse, err := c.restyReq.Delete(baseUrl + "/v1/vhosts/default/apps/app/streams/" + streamName)
	if err != nil {
		fmt.Println("Failed to delete stream:", err)
		return err
	}
	fmt.Printf("Delete stream response: %v\n", deleteResponse)
	if deleteResponse.IsError() && deleteResponse.StatusCode() != http.StatusNotFound {
		fmt.Println("Delete stream failed with status code:", deleteResponse.StatusCode(), deleteResponse.Request.URL)
		return fmt.Errorf("delete stream failed with status code: %d", deleteResponse.StatusCode())
	}
	return nil
}
