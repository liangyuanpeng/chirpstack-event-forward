package client

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	client = &http.Client{Timeout: time.Duration(30) * time.Second}
)

type DeviceQueueItem struct {
	Confirmed  bool
	Data       string
	DevEUI     string
	FCnt       int
	FPort      int
	JsonObject string
}

func DownLink(deviceQueueItem *DeviceQueueItem) error {
	//TODO token?
	//TODO url
	url := ""
	resp, err := client.Post(url+"/api/devices/"+deviceQueueItem.DevEUI+"/queue", "application/json", strings.NewReader("name=cjb"))
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
