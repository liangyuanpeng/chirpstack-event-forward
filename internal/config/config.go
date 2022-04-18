package config

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
}

type IntegrationsConfig struct {
	Mqtt MqttConfig `yaml:"mqtt"`
}

type MqttConfig struct {
	Enabled       bool   `yaml:"enabled"`
	TopicTemplate string `yaml:"topicTemplate"`
	Url           string `yaml:"url"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	ClientId      string `yaml:"clientId"`
	QOS           uint8  `yaml:"qos"`
}
