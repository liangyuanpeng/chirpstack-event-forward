package cmd

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	asintegration "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/liangyuanpeng/chirpstack-event-forward/internal/config"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration/mqtt"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration/pulsar"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

type EventData struct {
	ApplicationID string `json:"applicationID"`
	DevEUI        []byte `json:"devEUI"`
	FPort         int    `json:"fPort"`
	FCnt          int    `json:"fCnt"`
	DevAddr       []byte `json:devAddr`
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("&config.C.General.Http.Port:", config.C.General.Http.Port)
	listener := fmt.Sprintf(":%d", config.C.General.Http.Port)
	log.WithField("listener", listener).Info("started event forward!")
	http.Handle("/api", &handler{json: true})
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(listener, nil)
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

var integrations []integration.Integration

func initIntegration() {

	if config.C.Config[0].Integrations.Mqtt.Enabled {
		mqttI, err := mqtt.New(config.C.Config[0].Integrations.Mqtt)
		if err != nil {
			log.WithError(err).Fatalln("new mqtt integration failed!")
		}
		integrations = append(integrations, mqttI)
	}

	if config.C.Config[0].Integrations.Pulsar.Enabled {
		pulsarI, err := pulsar.New(config.C.Config[0].Integrations.Pulsar)
		if err == nil {
			integrations = append(integrations, pulsarI)
		} else {
			log.WithError(err).Fatalln("new pulsar integration failed!")
		}
	}

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
		log.WithError(err).Println("parse request body to proto failed!", "event", event)
		return
	}

	go func() {

		ch := make(chan integration.HandleError)

		for _, i := range integrations {
			name, err := i.HandleEvent(context.TODO(), ch, headerMap, b)
			if err != nil {
				log.WithError(err).Println("handle event failed!", "event", event, "name", name)
			}
		}

		for {
			handleErr := <-ch
			log.WithError(handleErr.Err).Println("handle event failed!", "name", handleErr.Name)
		}

	}()

}

func (h *handler) up(b []byte, vars map[string]string) error {
	var up asintegration.UplinkEvent
	var event EventData
	if err := h.unmarshal(b, &up, &event); err != nil {
		log.Println("unmarshal.up.failed!")
		return err
	}
	vars["appid"] = event.ApplicationID
	vars["devEUI"] = hex.EncodeToString(event.DevEUI)
	log.Println("Received up event!", "device", hex.EncodeToString(event.DevEUI))
	return nil
}

func (h *handler) join(b []byte, vars map[string]string) error {
	var join asintegration.JoinEvent
	var event EventData
	if err := h.unmarshal(b, &join, &event); err != nil {
		return err
	}
	vars["appid"] = event.ApplicationID
	vars["devEUI"] = hex.EncodeToString(event.DevEUI)
	log.Println("Received join event!", "device", hex.EncodeToString(event.DevEUI), "devaddr", hex.EncodeToString(event.DevAddr))
	return nil
}

func (h *handler) unmarshal(b []byte, v proto.Message, event *EventData) error {
	if h.json {
		if event != nil {
			return json.Unmarshal(b, event)
		} else {
			unmarshaler := &jsonpb.Unmarshaler{
				AllowUnknownFields: true, // we don't want to fail on unknown fields
			}
			return unmarshaler.Unmarshal(bytes.NewReader(b), v)
		}

	}
	return proto.Unmarshal(b, v)
}
