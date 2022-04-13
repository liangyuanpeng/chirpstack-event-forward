package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	asintegration "github.com/brocaar/chirpstack-api/go/v3/as/integration"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/liangyuanpeng/chirpstack-forward/internal/config"
	"github.com/liangyuanpeng/chirpstack-forward/internal/integration"
	"github.com/liangyuanpeng/chirpstack-forward/internal/integration/mqtt"
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to configuration file (optional)")
	rootCmd.PersistentFlags().Int("log-level", 4, "debug=5, info=4, error=2, fatal=1, panic=0")
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
		viper.AddConfigPath("$HOME/.config/cchirpstack-event-forward")
		viper.AddConfigPath("/etc/chirpstack-event-forward")
		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigFileNotFoundError:
				log.Warning("No configuration file found, using defaults. See: https://www.chirpstack.io/network-server/install/config/")
			default:
				log.WithError(err).Fatal("read configuration file error")
			}
		}
	}

	err := viper.Unmarshal(&config.C)
	if err != nil {
		panic(err)
	}
	fmt.Println("config.C:", config.C.Forwards)

	initIntegration()

}

var inte integration.Integration

func initIntegration() {
	mqttI, err := mqtt.New(config.C.Forwards[0].Integrations.Mqtt)
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

	headerMap := make(map[string]string)
	for name, values := range r.Header {
		// Loop over all values for the name.
		for _, value := range values {
			fmt.Println(name, value)
		}
		headerMap[name] = values[0]
	}
	headerMap["event"] = event
	headerMap["appid"] = ""
	headerMap["devEUI"] = ""

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.WithError(err).Println("read request body failed!")
		return
	}
	if len(b) < 1 {
		log.Error("have not request body!")
		return
	}
	var asevent asintegration.IntegrationEvent
	unmarshaler := &jsonpb.Unmarshaler{
		AllowUnknownFields: true, 
	}
	err = unmarshaler.Unmarshal(bytes.NewReader(b), &asevent)
	if err != nil {
		panic(err)
	}
	inte.HandleEvent(context.TODO(), b)
}
