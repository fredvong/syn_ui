package main

import (
	"syn_ui/common"
	"syn_ui/web"
)

func main() {
	var (
		err error
		config *common.Config
	)
	if config, err = common.LoadConfig("/Users/fvong/work/syn_ui/resource/example_config.json"); err != nil {
		panic(err)
	}
	web.InitWebServer(config)
}
