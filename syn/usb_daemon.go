package syn

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"github.com/sirupsen/logrus"
	"io"
	"log"
)

func RunCommand(port io.ReadWriteCloser, data []byte) (readback []byte, err error) {
	logrus.Debugf("Run command %v", data)
	if _, err = port.Write(data); err != nil {
		return
	}
	readback = make([]byte, len(data))
	var (
		n int
	)
	if n, err = port.Read(readback); err != nil {
		return
	}
	if n != len(data) {
		err = fmt.Errorf("%d is expected, actual is %d", len(data), n)
		return
	}
	for i, v := range readback {
		if v != data[i] {
			err = fmt.Errorf("expected value at byte %d should be %d but get %d", i, data[i], v)
			return
		}
	}
	return
}

func RunDaemon (portName string) (queue chan <- []byte, err error){
	var (
		port io.ReadWriteCloser
	)

	// Set up options.
	options := serial.OpenOptions{
		PortName: portName,
		BaudRate: 19200,
		DataBits: 8,
		StopBits: 1,
		MinimumReadSize: 1,
	}

	// Open the port.
	port, err = serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	myQueue := make(chan []byte)
	queue = myQueue

	go func() {
		// Make sure to close it later.
		defer port.Close()
		defer close(myQueue)
		for {
			select {
			case data := <- myQueue:
				if _, err1 := RunCommand(port, data); err1 != nil {
					logrus.Error(err)
				}
			}
		}
	}()

	return
}