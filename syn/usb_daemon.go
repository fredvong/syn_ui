package syn

import (
	"fmt"
	"github.com/jacobsa/go-serial/serial"
	"github.com/sirupsen/logrus"
	"io"
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
	logrus.Infof("Running USB daemon")
	var (
		port io.ReadWriteCloser
		myQueue = make(chan []byte)
	)

	queue = myQueue

	// Set up options.
	options := serial.OpenOptions{
		PortName: portName,
		BaudRate: 19200,
		DataBits: 8,
		StopBits: 1,
		MinimumReadSize: 1,
	}

	go func() {
		// Make sure to close it later.
		defer func() {
			if port != nil {
				port.Close()
			}
		}()
		defer close(myQueue)

		for {
			select {
			case data := <- myQueue:
				if port == nil {
					// open the port.
					if port, err = serial.Open(options); err != nil {
						logrus.Errorf("failed to connect to USB %v", err)
						if port != nil {
							port.Close()
							port = nil
						}
					}
				}
				if port != nil {
					if _, err = RunCommand(port, data); err != nil {
						logrus.Errorf("failed to run command %+v with error: %s", data, err)
						port.Close()
						port = nil
					}
				}
			}
		}
	}()

	return
}