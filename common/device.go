package common

import (
	"encoding/json"
	"io/ioutil"
)

type ButtonConfig struct {
	StartValue string   `json:"start_value"`
	Labels     []string `json:"labels"`
}

type KnobConfig struct {
	Labels []string `json:"labels"`
}

type Config struct {
	Name         string       `json:"name"`
	Device       string       `json:"device"`
	ButtonConfig ButtonConfig `json:"button_config"`
	KnobConfig   KnobConfig   `json:"knob_config"`
}

func LoadConfig(path string) (config *Config, err error) {
	var (
		data []byte
	)
	if data, err = ioutil.ReadFile(path); err != nil {
		return
	}
	config = &Config{}
	err = json.Unmarshal(data, config)
	return
}
