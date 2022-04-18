package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"

	asintegration "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/liangyuanpeng/chirpstack-event-forward/internal/config"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration/mqtt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "chirpstack-event-forward",
	Short: "ChirpStack Event Forward",
	Long: `ChirpStack Event Forward is an open-source Forward Server.
	> documentation & support: https://github.com/liangyuanpeng/chirpstack-event-forward
	> source & copyright information: https://github.com/liangyuanpeng/chirpstack-event-forward`,
	RunE: run,
}

func init() {
	cobra.OnInitialize(initConfig)
}

var cfgFile string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("&config.C.General.Http.Port:", config.C.General.Http.Port)
	listener := fmt.Sprintf(":%d", config.C.General.Http.Port)
	log.WithField("listener", listener).Info("started event forward!")
	err := http.ListenAndServe(listener, &handler{json: true})
	if err != nil {
		log.WithError(err).Fatal("start http server failed!")
	}
}

func initConfig() {

	cfgFile := ""

	if cfgFile != "" {
		b, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
		viper.SetConfigType("yaml")
		if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
			log.WithError(err).WithField("config", cfgFile).Fatal("error loading config file")
		}
	} else {
		viper.SetConfigName("chirpstack-event-forward")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/chirpstack-event-forward")
		viper.AddConfigPath("/etc/chirpstack-event-forward")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("No configuration file found, using defaults. ")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}

	err := viper.Unmarshal(&config.C)
	if err != nil {
		panic(err)
	}

	initIntegration()

}

var inte integration.Integration

func initIntegration() {
	mqttI, err := mqtt.New(config.C.Config[0].Integrations.Mqtt)
	if err != nil {
		log.WithError(err).Fatalln("new mqtt integration failed!")
	}
	inte = mqttI
}

type handler struct {
	json bool
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	event := r.URL.Query().Get("event")

	if event != "up" && event != "join" {
		log.WithField("event", event).Println("this event is not implemented")
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Println("read request body failed!")
		return
	}
	if len(b) < 1 {
		log.Error("have not request body!")
		return
	}

	headerMap := make(map[string]string)
	for name, values := range r.Header {
		headerMap[name] = values[0]
	}
	headerMap["event"] = event
	switch event {
	case "up":
		err = h.up(b, headerMap)
	case "join":
		err = h.join(b, headerMap)
	}
	if err != nil {
		log.WithError(err).Println("read request body failed!")
		return
	}

	inte.HandleEvent(context.TODO(), headerMap, b)
}

func (h *handler) up(b []byte, vars map[string]string) error {
	var up asintegration.UplinkEvent
	if err := h.unmarshal(b, &up); err != nil {
		return err
	}
	vars["appid"] = fmt.Sprintf("%d", up.ApplicationId)
	vars["devEUI"] = hex.EncodeToString(up.DevEui)
	log.Printf("Uplink received from %s with payload: %s", hex.EncodeToString(up.DevEui), hex.EncodeToString(up.Data))
	return nil
}

func (h *handler) join(b []byte, vars map[string]string) error {
	var join asintegration.JoinEvent
	if err := h.unmarshal(b, &join); err != nil {
		return err
	}
	vars["appid"] = fmt.Sprintf("%d", join.ApplicationId)
	vars["devEUI"] = hex.EncodeToString(join.DevEui)

	log.Printf("Device %s joined with DevAddr %s", hex.EncodeToString(join.DevEui), hex.EncodeToString(join.DevAddr))
	return nil
}

func (h *handler) unmarshal(b []byte, v proto.Message) error {
	if h.json {
		unmarshaler := &jsonpb.Unmarshaler{
			AllowUnknownFields: true, // we don't want to fail on unknown fields
		}
		return unmarshaler.Unmarshal(bytes.NewReader(b), v)
	}
	return proto.Unmarshal(b, v)
}
