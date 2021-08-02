package web

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"syn_ui/common"
	"syn_ui/syn"
)

var (
	uiConfig *common.Config
	queue    chan<- []byte
)

func hello(w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
}

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func mainPage(w http.ResponseWriter, reg *http.Request) {
	var (
		err  error
		data []byte
	)
	if data, err = ioutil.ReadFile("/Users/fvong/work/syn_ui/resource/main.html"); err != nil {
		return
	}

	fmt.Fprintf(w, string(data))
}

func getConfigure(w http.ResponseWriter, reg *http.Request) {
	var (
		data []byte
		err  error
	)
	if data, err = json.Marshal(uiConfig); err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(data))
}

func handleKeyEvent(w http.ResponseWriter, reg *http.Request) {
	if reg.Method == "PUT" {
		defer reg.Body.Close()
		data, _ := ioutil.ReadAll(reg.Body)

		log.Printf("body: %v", string(data))
		fmt.Fprintf(w, string(data))

		var (
			key string
			value string
			ok bool
			output = make([]byte, 3)
			KeyValue int64
			err error
			commandMap = make(map[string]string)
		)
		if err = json.Unmarshal(data, &commandMap); err != nil {
			logrus.Error(err)
			return
		}
		if key, ok = commandMap["key"]; !ok {
			logrus.Error("could not find key")
			return
		}
		if value, ok = commandMap["value"]; !ok {
			logrus.Error("could not find value")
			return
		}
		if value == "down" {
			output[0] = 0x90
		} else {
			output[0] = 0x80
		}
		if strings.HasPrefix(key, "0x") {
			key = strings.TrimPrefix(key, "0x")
		}
		if KeyValue, err = strconv.ParseInt(key, 16, 16); err != nil {
			logrus.Error(err)
			return
		}
		output[1] = byte(KeyValue)
		output[2] = 127
		queue <- output
	}
}

func InitWebServer(config *common.Config) {
	var (
		err error
	)

	uiConfig = config
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/config", getConfigure)
	http.HandleFunc("/key", handleKeyEvent)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/headers", headers)

	if queue, err = syn.RunDaemon("/dev/tty.usbmodem208A366847521"); err != nil {
		panic(err)
	}
	if err = http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}
