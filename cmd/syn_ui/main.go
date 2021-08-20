package main

import (
	"fmt"
	"os"
	"syn_ui/common"
	"syn_ui/web"
)

func main() {
	var (
		err error
		config *common.Config
		filename string
		args []string
	)
	args = os.Args[1:]
	if len(args) < 1 {
		fmt.Println("usage: syn_ui <config file>")
		return
	}
	filename = args[0]
	if len(filename) == 0 {
		panic("filename cannot be empty")
	}
	// filename = "/Users/fvong/work/syn_ui/resource/example_config.json"
	if config, err = common.LoadConfig(filename); err != nil {
		panic(err)
	}
	web.InitWebServer(config)
}
