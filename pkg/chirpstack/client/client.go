package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func DownLink(deviceQueueItem *DeviceQueueItem) error {
	url := ""
	token := ""
	song := make(map[string]interface{})
	song["deviceQueueItem"] = deviceQueueItem

	marshal, err := json.Marshal(song)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url+"/api/devices/"+deviceQueueItem.DevEUI+"/queue", bytes.NewReader(marshal))
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("Grpc-Metadata-Authorization", "Bearer "+token)

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
				// panic(err)
				log.Println("write.resp.content.failed!", resp.StatusCode, err)
				return err
			}
		}
		return errors.New(result.String())
	}
	log.Println("resp:", resp.StatusCode)
	return nil
}
