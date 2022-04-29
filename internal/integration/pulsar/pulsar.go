package pulsar

import (
	"bytes"
	"context"
	"sync"
	"text/template"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/config"
	log "github.com/sirupsen/logrus"
)

type Integration struct {
	topicTemplate *template.Template
	client        pulsar.Client
	producers     map[string]pulsar.Producer
	mutex         sync.Mutex
}

func New(config config.PulsarConfig) (*Integration, error) {

	t := template.New("Person template")
	tem, err := t.Parse(config.TopicTemplate)
	if err != nil {
		return nil, err
	}
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               config.Url,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	i := &Integration{
		client:        client,
		topicTemplate: tem,
		producers:     map[string]pulsar.Producer{},
	}
	return i, nil
}

func (i *Integration) HandleEvent(ctx context.Context, vars map[string]string, data []byte) (string, error) {

	buf := new(bytes.Buffer)
	i.topicTemplate.Execute(buf, vars)
	topic := string(buf.Bytes())
	log.Infof("integration/pulsar: topic: %s", topic)

	i.mutex.Lock()
	producer, ok := i.producers[topic]
	if ok {
		i.mutex.Unlock()
	} else {
		tmp, err := i.client.CreateProducer(pulsar.ProducerOptions{
			Topic: topic,
		})
		if err != nil {
			i.mutex.Unlock()
			return "pulsar", err
		}
		producer = tmp
		i.producers[topic] = tmp
		i.mutex.Unlock()
	}

	_, err := producer.Send(context.TODO(), &pulsar.ProducerMessage{
		Payload: data,
	})

	return "pulsar", err
}

func (i *Integration) Close() error {
	i.client.Close()
	return nil
}
