package syn

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"github.com/sirupsen/logrus"
	"io"
)

var (
	OKStatusString    = "OK!"
	StatusByteLength  = len(OKStatusString)
	ERRStatusString   = "ERR"
	MaximumBufferSize = 512
)

type CommandResponse struct {
	Data  []int64
	Error string
}

type CommandRequest struct {
	Command       []byte
	ResponseQueue chan<- CommandResponse
}

func RunCommand(port io.ReadWriteCloser, data []byte) (response []byte, err error) {
	var (
		n      int
		status string
	)
	response = make([]byte, MaximumBufferSize)
	if _, err = port.Write(data); err != nil {
		return
	}
	if n, err = port.Read(response); err != nil {
		return
	}
	if n < StatusByteLength {
		err = fmt.Errorf("incompleted status, expect %d bytes or more: %+v", StatusByteLength, response[:n])
		return
	}
	if status = string(response[n-StatusByteLength : n]); status != OKStatusString {
		err = fmt.Errorf("returned status not OK!, %+v", response[:n])
		return
	}
	response = response[:n-StatusByteLength]
	return
}

// RunDaemon runs the USB daemon to handle the MIDI like command syntax
func RunDaemon(portName string) (commandQueue chan<- CommandRequest, err error) {
	logrus.Infof("Running USB daemon")
	var (
		port           io.ReadWriteCloser
		myCommandQueue = make(chan CommandRequest)
	)

	commandQueue = myCommandQueue

	// Set up options.
	options := serial.OpenOptions{
		PortName:        portName,
		BaudRate:        19200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	}

	go func() {
		// Make sure to close it later.
		defer func() {
			var err error
			if port != nil {
				if err = port.Close(); err != nil {
					logrus.Errorf("failed to close port: %s", err)
				}
			}
		}()
		defer close(myCommandQueue)

		var response []byte
		for {
			select {
			case commandRequest := <-myCommandQueue:
				if port == nil {
					// open the port.
					if port, err = serial.Open(options); err != nil {
						logrus.Errorf("failed to connect to USB %v", err)
						if port != nil {
							if err = port.Close(); err != nil {
								logrus.Errorf("failed to close port: %s", err)
							}
							port = nil
						}
					}
				}
				if port != nil {
					if response, err = RunCommand(port, commandRequest.Command); err != nil {
						if err = port.Close(); err != nil {
							logrus.Errorf("failed to close port: %s", err)
						}
						port = nil
					}
				} else {
					err = fmt.Errorf("failed to connect to port %s", portName)
				}
				responseInt64 := make([]int64, len(response))
				for i, value := range response {
					responseInt64[i] = int64(value)
				}
				if err != nil {
					commandRequest.ResponseQueue <- CommandResponse{
						Data:  responseInt64,
						Error: err.Error(),
					}
				} else {
					commandRequest.ResponseQueue <- CommandResponse{
						Data:  responseInt64,
						Error: "",
					}
				}
			}
		}
	}()

	return
}
