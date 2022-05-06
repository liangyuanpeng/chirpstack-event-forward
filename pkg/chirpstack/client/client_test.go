package client

import "testing"

func Test_Downlink(t *testing.T) {
	err := DownLink(&DeviceQueueItem{
		Confirmed:  false,
		DevEUI:     "ffffff10000163cc",
		FCnt:       1,
		FPort:      2,
		JsonObject: "json",
		Data:       "MQ==",
	})
	if err != nil {
		panic(err)
	}
}
