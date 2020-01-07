package transport

import (
	"context"
)

type MockTransport struct {
	Ch      chan []byte
	Error   error
	History [][]byte
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		Ch: make(chan []byte),
	}
}

func (m *MockTransport) Run(context.Context) error {
	return m.Error
}

func (m *MockTransport) Receive() <-chan []byte {
	return m.Ch
}
