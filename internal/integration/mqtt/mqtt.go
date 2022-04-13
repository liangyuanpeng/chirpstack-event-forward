package mqtt

import (
	"context"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangyuanpeng/chirpstack-forward/internal/config"
	log "github.com/sirupsen/logrus"
)

// Integration implements an Mqtt integration.
type Integration struct {
	conn   mqtt.Client
	config config.MqttConfig
	topic  string
}

func New(config config.MqttConfig) (*Integration, error) {

	i := &Integration{
		config: config,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Url)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetClientID(config.ClientId)

	i.conn = mqtt.NewClient(opts)
	for {
		if token := i.conn.Connect(); token.Wait() && token.Error() != nil {
			log.Errorf("integration/mqtt: connecting to broker error, will retry in 2s: %s", token.Error())
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	return i, nil
}

func (i *Integration) HandleEvent(ctx context.Context, data []byte) error {
	if token := i.conn.Publish(i.topic, i.config.QOS, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (i *Integration) Close() error {
	if token := i.conn.Unsubscribe(i.topic); token.Wait() && token.Error() != nil {
		return fmt.Errorf("integration/mqtt: unsubscribe from %s error: %s", i.topic, token.Error())
	}
	return nil
}
