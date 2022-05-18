package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	client = &http.Client{Timeout: time.Duration(30) * time.Second}
)

type DeviceQueueItem struct {
	Confirmed  bool   `json:"confirmed"`
	Data       string `json:"data"`
	DevEUI     string `json:"devEUI"`
	FCnt       int    `json:"fCnt"`
	FPort      int    `json:"fPort"`
	JsonObject string `json:"jsonObject"`
}

type ChirpstackClient struct {
	Url   string
	Token string
}

func New(url, token string) (*ChirpstackClient, error) {
	if url == "" || token == "" {
		return nil, errors.New("url or apitoken is empty")
	}
	return &ChirpstackClient{
		Url:   url,
		Token: token,
	}, nil
}

func (c *ChirpstackClient) DownLink(ctx context.Context, deviceQueueItem *DeviceQueueItem) error {

	if deviceQueueItem.DevEUI == "" {
		return errors.New("downlink must be set devEUI")
	}

	reqdata := make(map[string]interface{})
	reqdata["deviceQueueItem"] = deviceQueueItem

	log.Println("Received downlink event!", "device", deviceQueueItem.DevEUI)

	marshal, err := json.Marshal(reqdata)
	if err != nil {
		return err
	}

	// simple retry once
	if c.sendDownlinkRequest(ctx, deviceQueueItem.DevEUI, marshal) != nil {
		return c.sendDownlinkRequest(ctx, deviceQueueItem.DevEUI, marshal)
	}

	return nil
}

func (c *ChirpstackClient) sendDownlinkRequest(ctx context.Context, devEUI string, data []byte) error {

	request, err := http.NewRequestWithContext(ctx, "POST", c.Url+"/api/devices/"+devEUI+"/queue", bytes.NewReader(data))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("Grpc-Metadata-Authorization", "Bearer "+c.Token)

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var buffer [512]byte
		result := bytes.NewBuffer(nil)
		for {
			n, err := resp.Body.Read(buffer[0:])
			result.Write(buffer[0:n])
			if err != nil && err == io.EOF {
				break
			} else if err != nil {
				log.Println("write.resp.content.failed!", resp.StatusCode, err)
				return err
			}
		}
		return errors.New(result.String())
	}
	log.Println("sendDownlinkRequest", "statusCode", resp.StatusCode)
	return nil
}
