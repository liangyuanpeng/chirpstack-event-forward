package config

import (
	"github.com/liangyuanpeng/chirpstack-event-forward/internal/integration"
	"github.com/liangyuanpeng/chirpstack-event-forward/pkg/chirpstack/client"
)

var C RootConfig

type RootConfig struct {
	Version string

	General GeneralConfig `yaml:"general"`

	Config []ForwardConfig `yaml:"config"`
}

type GeneralConfig struct {
	Http HTTPConfig `yaml:"http"`
}

type HTTPConfig struct {
	Port int `yaml:"port"`
}

type ForwardConfig struct {
	Name             string             `yaml:"name"`
	ChirpstackConfig ChirpstackConfig   `yaml:"chirpstack"`
	Integrations     IntegrationsConfig `yaml:"integrations"`
}

type ChirpstackConfig struct {
	Account  string `yaml:"account"`
	Password string `yaml:"password"`
	Url      string `yaml:"url"`
	ApiToken string `yaml:"apiToken"`
}

type IntegrationsConfig struct {
	Mqtt   MqttConfig   `yaml:"mqtt"`
	Pulsar PulsarConfig `yaml:"pulsar"`
}

type MqttConfig struct {
	Enabled       bool   `yaml:"enabled"`
	TopicTemplate string `yaml:"topicTemplate"`
	Url           string `yaml:"url"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	ClientId      string `yaml:"clientId"`
	QOS           uint8  `yaml:"qos"`
	DownlinkTopic string `yaml:"downlinkTopic"`
}

type PulsarConfig struct {
	Enabled              bool   `yaml:"enabled"`
	TopicTemplate        string `yaml:"topicTemplate"`
	Url                  string `yaml:"url"`
	ProducerNameTemplate string `yaml:"producerNameTemplate"`
	TopicsPattern        string `yaml:"topicsPattern"`
	SubscriptionName     string `yaml:"subscriptionName"`
	ConsumerName         string `yaml:"consumerName"`
}

type IntegrationOption struct {
	ChirpstackClient *client.ChirpstackClient
	Ch               chan integration.HandleError
}
