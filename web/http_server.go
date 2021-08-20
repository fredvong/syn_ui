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
	velocity = byte(127)
	debug = false
)

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func mainPage(w http.ResponseWriter, reg *http.Request) {
	if !debug {
		fmt.Fprintf(w, mainPageHTML)
		return
	}

	// Running the debug mode.
	// Read html file from the main.html
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
			key        string
			value      string
			ok         bool
			output     = make([]byte, 3)
			KeyValue   int64
			err        error
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
		if KeyValue, err = strconv.ParseInt(key, 10, 16); err != nil {
			logrus.Error(err)
			return
		}
		output[1] = byte(KeyValue)
		output[2] = velocity
		queue <- output
	}
}

func handleControlEvent(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "PUT" {
		defer request.Body.Close()
		data, _ := ioutil.ReadAll(request.Body)

		log.Printf("body: %v", string(data))
		fmt.Fprintf(writer, string(data))

		var (
			controlValue      string
			value             string
			ok                bool
			output            = make([]byte, 3)
			controlValueInt64 int64
			valueInt64        int64
			err               error
			commandMap        = make(map[string]string)
		)
		if err = json.Unmarshal(data, &commandMap); err != nil {
			logrus.Error(err)
			return
		}
		if controlValue, ok = commandMap["control_value"]; !ok {
			logrus.Error("could not find control_value")
			return
		}
		if value, ok = commandMap["value"]; !ok {
			logrus.Error("could not find value")
			return
		}
		if controlValueInt64, err = strconv.ParseInt(controlValue, 10, 16); err != nil {
			logrus.Error(err)
			return
		}
		if valueInt64, err = strconv.ParseInt(value, 10, 16); err != nil {
			logrus.Error(err)
			return
		}

		output[0] = 176
		output[1] = byte(controlValueInt64)
		output[2] = byte(valueInt64)
		queue <- output
	}
}

func handleVelocityEvent(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "PUT" {
		defer request.Body.Close()
		data, _ := ioutil.ReadAll(request.Body)

		logrus.Infof("velocity: %v", string(data))
		fmt.Fprintf(writer, string(data))

		var (
			commandMap = make(map[string]string)
			err        error
			value      string
			ok         bool
			valueInt64 int64
		)
		if err = json.Unmarshal(data, &commandMap); err != nil {
			logrus.Error(err)
			return
		}
		if value, ok = commandMap["value"]; !ok {
			logrus.Error("could not find value")
			return
		}
		if valueInt64, err = strconv.ParseInt(value, 10, 16); err != nil {
			logrus.Error(err)
			return
		}
		velocity = byte(valueInt64);
	}
}

func InitWebServer(config *common.Config) {
	logrus.Infof("Running the http server")
	var (
		err error
	)

	uiConfig = config
	http.HandleFunc("/", mainPage)
	http.HandleFunc("/config", getConfigure)
	http.HandleFunc("/key", handleKeyEvent)
	http.HandleFunc("/control", handleControlEvent)
	http.HandleFunc("/velocity", handleVelocityEvent)
	http.HandleFunc("/headers", headers)

	if queue, err = syn.RunDaemon(config.Device); err != nil {
		logrus.Error(err)
	}
	if err = http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}
