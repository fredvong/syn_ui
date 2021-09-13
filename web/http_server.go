package web

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	DefaultPort = 8090
)

func headers(w http.ResponseWriter, req *http.Request) {
	var err error
	for name, headers := range req.Header {
		for _, h := range headers {
			if _, err = fmt.Fprintf(w, "%v: %v\n", name, h); err != nil {
				logrus.Error(err)
			}
		}
	}
}

func mainPage(w http.ResponseWriter, reg *http.Request) {
	var err error
	if !debug {
		if _, err = fmt.Fprintf(w, mainPageHTML); err != nil {
			logrus.Errorf("failed to render main page: %s", err)
		}
		return
	}

	// Running the debug mode.
	// Read html file from the main.html
	var data []byte
	if data, err = ioutil.ReadFile("/Users/fvong/work/syn_ui/resource/main.html"); err != nil {
		return
	}

	if _, err = fmt.Fprintf(w, string(data)); err != nil {
		logrus.Errorf("failed to render main page: %s", err)
	}
}

func getConfigure(w http.ResponseWriter, reg *http.Request) {
	var (
		data []byte
		err  error
	)
	if data, err = json.Marshal(uiConfig); err != nil {
		panic(err)
	}
	if _, err = fmt.Fprintf(w, string(data)); err != nil {
		logrus.Errorf("failed to render getConfigure: %s", err)
	}
}

func handleKeyEvent(w http.ResponseWriter, reg *http.Request) {
	var err error
	if reg.Method == "PUT" {
		defer func() {
			err = reg.Body.Close()
			if err != nil {
				logrus.Errorf("failed to close the request body in handleKeyEvent: %s", err)
			}
		}()
		data, _ := ioutil.ReadAll(reg.Body)

		log.Printf("body: %v", string(data))
		if _, err = fmt.Fprintf(w, string(data)); err != nil {
			logrus.Errorf("failed to render handleKeyEvent: %s", err)
		}

		var (
			key        string
			value      string
			ok         bool
			output     = make([]byte, 3)
			KeyValue   int64
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
	var err error
	if request.Method == "PUT" {
		var data []byte
		defer func() {
			var err error
			if err = request.Body.Close(); err != nil {
				logrus.Errorf("failed to close request body in handleControlEvent: %s", err)
			}
		}()
		if data, err = ioutil.ReadAll(request.Body); err != nil {
			logrus.Errorf("failed to read request body in handleControlEvent: %s", err)
		}

		logrus.Debugln("body: %v", string(data))
		if _, err = fmt.Fprintf(writer, string(data)); err != nil {
			logrus.Errorf("failed to render handleControlEvent: %s", err)
		}

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
	var err error
	if request.Method == "PUT" {
		defer func() {
			var err error
			if err = request.Body.Close(); err != nil {
				logrus.Errorf("failed to close request body of handleVelocityEvent: %s", err)
			}
		}()
		data, _ := ioutil.ReadAll(request.Body)

		logrus.Infof("velocity: %v", string(data))
		if _, err = fmt.Fprintf(writer, string(data)); err != nil {
			logrus.Errorf("failed to redner handleVelocityEvent: %s", err)
		}

		var (
			commandMap = make(map[string]string)
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
		velocity = byte(valueInt64)
	}
}

func usbDevice(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		_, err := os.Stat(uiConfig.Device)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintf(writer, "{\"status\": \"device %s not found\"}", uiConfig.Device)
			} else {
				fmt.Fprintf(writer, "{\"status\": \"device %s has error %s\"}", uiConfig.Device, err)
			}
			return
		}
		fmt.Fprintf(writer, "{\"status\": \"device %s found\"}", uiConfig.Device)
		return
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
	http.HandleFunc("/usb_device", usbDevice)

	if queue, err = syn.RunDaemon(config.Device); err != nil {
		logrus.Error(err)
	}
	logrus.Infof("listerning on port: %d", DefaultPort)
	if err = http.ListenAndServe(fmt.Sprintf(":%d", DefaultPort), nil); err != nil {
		panic(err)
	}
}

