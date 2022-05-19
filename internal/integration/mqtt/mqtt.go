package mqtt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/config"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration"
	"github.com/liangyuanpeng/chirpstack-event-forward/pkg/chirpstack/client"
	log "github.com/sirupsen/logrus"
)

// Integration implements an Mqtt integration.
type Integration struct {
	conn             mqtt.Client
	config           config.MqttConfig
	topicTemplate    *template.Template
	chirpstackClient *client.ChirpstackClient

	ch chan integration.HandleError
}

func New(config config.MqttConfig, opt *config.IntegrationOption) (*Integration, error) {

	chirpstackClient := opt.ChirpstackClient

	if config.Enabled && config.Url == "" {
		return nil, errors.New("integration/mqtt: empty url|")
	}

	t := template.New("mqtt topic template")
	tem, err := t.Parse(config.TopicTemplate)
	if err != nil {
		return nil, err
	}

	i := &Integration{
		config:           config,
		topicTemplate:    tem,
		chirpstackClient: chirpstackClient,
		ch:               opt.Ch,
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.Url)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetClientID(config.ClientId)
	opts.SetDefaultPublishHandler(i.messagePubHandler)
	opts.SetAutoReconnect(true)

	i.conn = mqtt.NewClient(opts)
	for {
		log.Info("integration/mqtt: connecting to broker")
		if token := i.conn.Connect(); token.Wait() && token.Error() != nil {
			log.Errorf("integration/mqtt: connecting to broker error, will retry in 2s: %s", token.Error())
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
	log.Info("integration/mqtt: connected to broker")

	if i.config.DownlinkTopic != "" {
		log.Infof("integration/mqtt: subscribing to broker :%s", i.config.DownlinkTopic)
		if token := i.conn.Subscribe(i.config.DownlinkTopic, 1, nil); token.Wait() && token.Error() != nil {
			log.Errorf("integration/mqtt: subscribing to broker error: %s", token.Error())
			return i, token.Error()
		}
	}
	log.Info("integration/mqtt: subscribed to broker")

	return i, nil
}

func (i *Integration) messagePubHandler(mqttclient mqtt.Client, msg mqtt.Message) {
	log.Println("received downlink event!", "payload", string(msg.Payload()), "topic", msg.Topic())
	if i.chirpstackClient != nil {
		dqi := &client.DeviceQueueItem{}
		err := json.Unmarshal(msg.Payload(), dqi)
		if err != nil {
			i.ch <- integration.HandleError{
				Err:  err,
				Name: "mqtt",
			}
		} else {
			err = i.chirpstackClient.DownLink(context.TODO(), dqi)
			if err != nil {
				i.ch <- integration.HandleError{
					Err:  err,
					Name: "mqtt",
				}
			}
		}
	}
}

func (i *Integration) HandleEvent(ctx context.Context, ch chan integration.HandleError, vars map[string]string, data []byte) (string, error) {

	topic := new(bytes.Buffer)
	i.topicTemplate.Execute(topic, vars)

	log.Infof("integration/mqtt: topic: %s", topic.Bytes())

	if token := i.conn.Publish(topic.String(), i.config.QOS, false, data); token.Wait() && token.Error() != nil {
		return "mqtt", token.Error()
	}
	return "mqtt", nil
}

func (i *Integration) Close() error {
	i.conn.Disconnect(1000)
	return nil
}
