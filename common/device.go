package common

import (
	"encoding/json"
	"io/ioutil"
)

// Button
// {
//     command: 0x90,
//     key: 0x3C,
//     velocity: 0x7F],
// }
type Button struct {
	Name       string `json:"name"`
	Key        string `json:"key"`
}

type Config struct {
	Name    string   `json:"name"`
	Buttons []Button `json:"buttons"`
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
