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
	// Configuration parameters
	Listen        string
	MaxPacketSize int
	Gateway       string

	ch                     chan []byte
	enabled                bool
	resolvedGatewayAddress *net.UDPAddr
	socket                 net.PacketConn
}

// NewUDP creates new instance of UDP transport
func NewUDP(cfg interface{}) (transport.Transport, error) {
	udp := &UDP{
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
	if udp.MaxPacketSize == 0 {
		udp.MaxPacketSize = 1024
	}
	udp.enabled = true

	// Resolve LoRa gateway address, if any
	if udp.Gateway != "" {
		udp.resolvedGatewayAddress, err = net.ResolveUDPAddr("udp", udp.Gateway)
		if err != nil {
			return nil, err
		}
	} else {
		glog.Info("UDP gateway address unset, packets will not be delivered to devices back")
	}

	return udp, err
}

// Run runs UDP server in blocking mode
func (r *UDP) Run(ctx context.Context) error {
	if !r.enabled {
		// If server is not enabled - simply blocks until context done
		glog.Info("UDP is not enabled")
		<-ctx.Done()
		return nil
	}

	// Create UDP listening socket
	var err error
	r.socket, err = net.ListenPacket("udp", r.Listen)
	if err != nil {
		return err
	}
	glog.Infof("UDP server started at %s", r.Listen)

	// Start receiver
	go r.serve(ctx)

	// Wait until context canceled
	<-ctx.Done()
	r.socket.Close()

	return nil
}

// Receive returns channel where UDP will send received packets to.
func (r *UDP) Receive() <-chan []byte {
	return r.ch
}

// Send sends packet to IOT device
func (r *UDP) Send(packet []byte) error {
	// Just do nothing in bypass mode
	if !r.enabled || r.resolvedGatewayAddress == nil {
		return nil
	}

	_, err := r.socket.WriteTo(packet, r.resolvedGatewayAddress)
	if err == nil {
		glog.Infof("UDP LoRa gateway: %v <-- %d bytes", r.resolvedGatewayAddress, len(packet))
	}

	return err
}

func (r *UDP) serve(ctx context.Context) {
	buf := make([]byte, r.MaxPacketSize)
	for {
		n, _, err := r.socket.ReadFrom(buf)
		if err != nil {
			// Terminate goroutine when listener closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			glog.Infof("readFrom failed: %v", err)
			continue
		}
		r.ch <- buf[:n]
	}
}
