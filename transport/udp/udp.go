package udp

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/mitchellh/mapstructure"

	"github.com/open-iot-devices/server/transport"
)

// UDP implements server/transport interface
type UDP struct {
	transport.BaseTransport

	// Configuration parameters
	Listen  string
	Address string

	name            string
	ch              chan []byte
	enabled         bool
	resolvedAddress *net.UDPAddr
	socket          net.PacketConn
}

// NewUDP creates new instance of UDP transport
func NewUDP(name string, cfg interface{}) (transport.Transport, error) {
	udp := &UDP{
		BaseTransport: transport.BaseTransport{
			Name: name,
		},
		ch: make(chan []byte, 1),
	}

	// If no configuration present - bypass mode
	if cfg == nil {
		return udp, nil
	}

	// Map / verify configuration
	err := mapstructure.Decode(cfg, udp)
	if udp.Listen == "" {
		return nil, errors.New("config parameter udp.listen is required")
	}

	// Resolve LoRa gateway address, if any
	if udp.Address != "" {
		udp.resolvedAddress, err = net.ResolveUDPAddr("udp", udp.Address)
		if err != nil {
			return nil, err
		}
	} else {
		glog.Info("UDP gateway address unset, packets will not be delivered to devices back")
	}

	return udp, err
}

// Start starts UDP server, non blocking call
func (r *UDP) Start(ctx context.Context) error {
	// Create UDP listening socket
	var err error
	r.socket, err = net.ListenPacket("udp", r.Listen)
	if err != nil {
		return err
	}
	glog.Infof("UDP server started at %s", r.Listen)

	// Start receiver
	go r.serve(ctx)

	// Start context listener:
	// since UDP read is blocking call, but, can be terminated
	// when socket is closed - starting one more goroutine
	// just to close socket once server terminated
	go func() {
		<-ctx.Done()
		r.socket.Close()
	}()

	return nil
}

// Receive returns channel where UDP will send received packets to.
func (r *UDP) Receive() <-chan []byte {
	return r.ch
}

// Send simply sends payload as UDP packet to address from configuration
func (r *UDP) Send(packet []byte) error {
	sent, err := r.socket.WriteTo(packet, r.resolvedAddress)
	if err != nil {
		return err
	}
	if sent != len(packet) {
		glog.Warningf("Transport %s: packet truncated! Sent only %d of %d",
			r.GetName(), sent, len(packet))
	}

	return nil
}

func (r *UDP) serve(ctx context.Context) {
	buf := make([]byte, 65535)

	for {
		n, _, err := r.socket.ReadFrom(buf)
		if err != nil {
			// Terminate goroutine when listener closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				glog.Info("terminated udp")
				return
			}
			glog.Infof("%s: readFrom failed: %v", r.GetName(), err)
			continue
		}
		r.ch <- buf[:n]
	}
}
