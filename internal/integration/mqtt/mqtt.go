package mqtt

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/config"
	log "github.com/sirupsen/logrus"
)

// Integration implements an Mqtt integration.
type Integration struct {
	conn   mqtt.Client
	config config.MqttConfig
	topic  string
}

func New(config config.MqttConfig) (*Integration, error) {

	if config.Enabled && config.Url == "" {
		return nil, errors.New("integration/mqtt: empty url|")
	}

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

func (i *Integration) HandleEvent(ctx context.Context, vars map[string]string, data []byte) error {
	t := template.New("Person template")
	tem, err := t.Parse(i.config.TopicTemplate)
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	tem.Execute(buf, vars)
	log.Println("topic.is:", string(buf.Bytes()))

	if token := i.conn.Publish(i.topic, i.config.QOS, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (i *Integration) Close() error {
	i.conn.Disconnect(1000)
	return nil
}
