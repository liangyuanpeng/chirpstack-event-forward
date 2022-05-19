package client

import (
	"context"
	"testing"
)

func Test_Downlink(t *testing.T) {

	c, err := New("", "")
	if err != nil {
		panic(err)
	}

	err = c.DownLink(context.TODO(), &DeviceQueueItem{
		Confirmed: false,
		DevEUI:    "ffffff100003717d",
		FPort:     36,
		Data:      "MQ==",
	})
	if err != nil {
		panic(err)
	}
}
