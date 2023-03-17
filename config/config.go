package config

import (
	"errors"
	"os"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Username    string
	ChatMessage []ChatMessageConfig `yaml:"chat_message" mapstructure:"chat_message"`
}

type ChatMessageConfig struct {
	Key      []string `mapstructure:"key"`
	Message  []string `mapstructure:"message"`
	Duration float64  `mapstructure:"duration"`
}

func NewConfig(logger *zap.Logger) (*Config, error) {
	f, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	var c Config
	var raw interface{}
	if err := yaml.Unmarshal(f, &raw); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{WeaklyTypedInput: true, Result: &c})
	if err := decoder.Decode(raw); err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	logger.Sugar().Debugf("%+v\n", c)

	if c.Username == "" {
		return nil, errors.New("Missing username in configuration")
	}

	return &c, nil
}
