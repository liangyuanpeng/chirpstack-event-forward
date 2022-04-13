package client

import "testing"

func Test_Downlink(t *testing.T) {
	err := DownLink(&DeviceQueueItem{
		Confirmed:  false,
		DevEUI:     "abc",
		FCnt:       1,
		FPort:      2,
		JsonObject: "json",
	})
	if err != nil {
		panic(err)
	}
}
