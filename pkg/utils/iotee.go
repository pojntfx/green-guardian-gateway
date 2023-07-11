package utils

import (
	"time"

	"gitlab.mi.hdm-stuttgart.de/iotee/go-iotee"
)

type IoTee interface {
	Open() error
	Close()
	RxPump()
	RxChan() chan *iotee.Message
	ReceiveWithTimeout(timeout time.Duration) *iotee.Message
	ReceiveBlocking() *iotee.Message
	Transmit(msg *iotee.Message) error
}

type IoTeeAdapter struct {
	original *iotee.IoTee
}

func NewIoTeeAdapter(original *iotee.IoTee) IoTee {
	return &IoTeeAdapter{original: original}
}

func (adapter *IoTeeAdapter) Open() error {
	return adapter.original.Open()
}

func (adapter *IoTeeAdapter) Close() {
	adapter.original.Close()
}

func (adapter *IoTeeAdapter) RxPump() {
	adapter.original.RxPump()
}

func (adapter *IoTeeAdapter) RxChan() chan *iotee.Message {
	return adapter.original.RxChan
}

func (adapter *IoTeeAdapter) ReceiveWithTimeout(timeout time.Duration) *iotee.Message {
	return adapter.original.ReceiveWithTimeout(timeout)
}

func (adapter *IoTeeAdapter) ReceiveBlocking() *iotee.Message {
	return adapter.original.ReceiveBlocking()
}

func (adapter *IoTeeAdapter) Transmit(msg *iotee.Message) error {
	return adapter.original.Transmit(msg)
}
