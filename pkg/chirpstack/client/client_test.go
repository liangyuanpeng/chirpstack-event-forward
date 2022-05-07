package client

import "testing"

func Test_Downlink(t *testing.T) {

	c, err := New("", "")
	if err != nil {
		panic(err)
	}

	err = c.downLink(&DeviceQueueItem{
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
