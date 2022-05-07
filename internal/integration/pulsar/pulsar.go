package pulsar

import (
	"bytes"
	"context"
	"sync"
	"text/template"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/config"
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration"
	"github.com/liangyuanpeng/chirpstack-event-forward/pkg/chirpstack/client"
	log "github.com/sirupsen/logrus"
)

type Integration struct {
	topicTemplate        *template.Template
	client               pulsar.Client
	producers            sync.Map
	producerNameTemplate *template.Template
	chirpstackClient     *client.ChirpstackClient
}

type ProducerGroup struct {
	producer pulsar.Producer
	sync.Mutex
}

func New(config config.PulsarConfig, chirpstackClient *client.ChirpstackClient) (*Integration, error) {

	t := template.New("Person template")
	tem, err := t.Parse(config.TopicTemplate)
	if err != nil {
		return nil, err
	}

	t2 := template.New("pulsar producer name template")
	tem2, err2 := t2.Parse(config.ProducerNameTemplate)
	if err2 != nil {
		return nil, err2
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
		client:               client,
		topicTemplate:        tem,
		producers:            sync.Map{},
		producerNameTemplate: tem2,
		chirpstackClient:     chirpstackClient,
	}
	return i, nil
}

func (i *Integration) HandleEvent(ctx context.Context, ch chan integration.HandleError, vars map[string]string, data []byte) (string, error) {

	buf := new(bytes.Buffer)
	i.topicTemplate.Execute(buf, vars)
	topic := buf.String()

	buf2 := new(bytes.Buffer)
	i.producerNameTemplate.Execute(buf2, vars)
	produerName := buf.String()

	log.Infof("integration/pulsar: topic: %s", topic)
	key := produerName

	pgt, _ := i.producers.LoadOrStore(key, &ProducerGroup{})
	pg := pgt.(*ProducerGroup)

	if pg.producer == nil {
		pg.Lock()
		if pg.producer == nil {
			tmp, err := i.client.CreateProducer(pulsar.ProducerOptions{
				Topic: topic,
				Name:  produerName,
			})
			if err != nil {
				pg.Unlock()
				return "pulsar", err
			}
			pg.producer = tmp
		}
		pg.Unlock()
	}

	pg.producer.SendAsync(context.TODO(), &pulsar.ProducerMessage{
		Payload: data,
	}, func(mi pulsar.MessageID, pm *pulsar.ProducerMessage, err error) {
		if err != nil {
			ch <- integration.HandleError{
				Err:  err,
				Name: "pulsar",
			}
		}

	})

	// _, err := pg.producer.Send(context.TODO(), &pulsar.ProducerMessage{
	// 	Payload: data,
	// })

	return "pulsar", nil
}

func (i *Integration) Close() error {
	i.client.Close()
	return nil
}
