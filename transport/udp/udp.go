package udp

import (
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
	Listen string
	Remote string

	name            string
	receiveCh       chan []byte
	resolvedAddress *net.UDPAddr
	socket          net.PacketConn
}

// NewUDP creates new instance of UDP transport
func NewUDP(name string, cfg interface{}) (transport.Transport, error) {
	udp := &UDP{
		BaseTransport: transport.BaseTransport{
			Name: name,
		},
	}

	// If no configuration present - bypass mode
	if cfg == nil {
		return udp, nil
	}

	// Map / verify configuration
	err := mapstructure.Decode(cfg, udp)
	if udp.Listen == "" {
		return nil, errors.New("config parameter udp.ListenAddress is required")
	}

	// Resolve LoRa gateway address, if any
	if udp.Remote != "" {
		udp.resolvedAddress, err = net.ResolveUDPAddr("udp", udp.Remote)
		if err != nil {
			return nil, err
		}
	} else {
		glog.Info("UDP gateway address unset, packets will not be delivered to devices back")
	}

	return udp, err
}

// Start starts UDP server, non blocking call
func (r *UDP) Start() error {
	// Create UDP listening socket
	var err error
	if r.socket, err = net.ListenPacket("udp", r.Listen); err != nil {
		return err
	}
	glog.Infof("%s: UDP server started at %s", r.GetName(), r.Listen)

	// Start UDP listener (with capability of buffer one packet)
	r.receiveCh = make(chan []byte, 1)
	go r.serve()

	return nil
}

// Stop performs graceful shutdown of UDP server
func (r *UDP) Stop() {
	// Goroutine will be canceled once socket.ReadFrom returns with error
	r.socket.Close()
}

// Receive returns channel where UDP will send received packets to.
func (r *UDP) Receive() <-chan []byte {
	return r.receiveCh
}

// Send simply sends payload as UDP packet to address from configuration
func (r *UDP) Send(packet []byte) error {
	sent, err := r.socket.WriteTo(packet, r.resolvedAddress)
	if err != nil {
		return err
	}
	if sent != len(packet) {
		glog.Warningf("%s: packet truncated! Sent only %d of %d",
			r.GetName(), sent, len(packet))
	}

	return nil
}

func (r *UDP) serve() {
	buf := make([]byte, 65535)

	for {
		// ReadFrom is blocking call unless socket closed
		n, _, err := r.socket.ReadFrom(buf)
		if err != nil {
			// Terminate goroutine if socket closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			glog.Infof("%s: readFrom failed: %v", r.GetName(), err)
			continue
		}
		r.receiveCh <- buf[:n]
	}
}
