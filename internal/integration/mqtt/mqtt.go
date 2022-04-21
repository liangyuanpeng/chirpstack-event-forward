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
	conn          mqtt.Client
	config        config.MqttConfig
	topic         string
	topicTemplate *template.Template
}

func New(config config.MqttConfig) (*Integration, error) {

	if config.Enabled && config.Url == "" {
		return nil, errors.New("integration/mqtt: empty url|")
	}

	t := template.New("Person template")
	tem, err := t.Parse(config.TopicTemplate)
	if err != nil {
		return nil, err
	}

	i := &Integration{
		config:        config,
		topicTemplate: tem,
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

	buf := new(bytes.Buffer)
	i.topicTemplate.Execute(buf, vars)

	log.Infof("integration/mqtt: topic: %s", buf.Bytes())

	if token := i.conn.Publish(i.topic, i.config.QOS, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (i *Integration) Close() error {
	i.conn.Disconnect(1000)
	return nil
}
