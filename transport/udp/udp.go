package udp

import (
	"errors"
	"net"
	"strings"

	"github.com/golang/glog"
	"github.com/open-iot-devices/server/transport"
)

const typeName = "udp"

// UDP implements server/transport interface
type UDP struct {
	Listen string
	Remote string

	name            string
	receiveCh       chan []byte
	resolvedAddress *net.UDPAddr
	socket          net.PacketConn
}

// NewUDP creates new instance of UDP transport
func NewUDP(name string) transport.Transport {
	return &UDP{
		name: name,
	}
}

// GetName returns transport name
func (s *UDP) GetName() string {
	return s.name
}

// GetTypeName returns type name of transport
func (s *UDP) GetTypeName() string {
	return typeName
}

// Start starts UDP server in background mode
func (s *UDP) Start() error {
	if s.Listen == "" {
		return errors.New("ListenAddress parameter required")
	}
	// Resolve LoRa gateway address, if any
	if s.Remote != "" {
		if addr, err := net.ResolveUDPAddr("udp", s.Remote); err != nil {
			s.resolvedAddress = addr
		} else {
			return err
		}
	} else {
		glog.Info("UDP gateway address unset, packets will not be delivered to devices back")
	}
	// Create UDP listening socket
	sock, err := net.ListenPacket("udp", s.Listen)
	if err != nil {
		return err
	}
	s.socket = sock
	glog.Infof("UDP server started at %s", sock)

	// Start UDP listener (with capability of buffer one packet)
	s.receiveCh = make(chan []byte, 1)
	go s.serve()

	return nil
}

// Stop performs graceful shutdown of UDP server
func (s *UDP) Stop() {
	// Goroutine will be canceled once socket.ReadFrom returns with error
	s.socket.Close()
}

// Receive returns channel where UDP will send received packets to.
func (s *UDP) Receive() <-chan []byte {
	return s.receiveCh
}

// Send simply sends payload as UDP packet to address from configuration
func (s *UDP) Send(packet []byte) error {
	sent, err := s.socket.WriteTo(packet, s.resolvedAddress)
	if err != nil {
		return err
	}
	if sent != len(packet) {
		glog.Warningf("packet truncated, sent only %d of %d",
			sent, len(packet))
	}

	return nil
}

func (s *UDP) serve() {
	buf := make([]byte, 65535)

	for {
		// ReadFrom is blocking call unless socket closed
		n, _, err := s.socket.ReadFrom(buf)
		if err != nil {
			// Terminate goroutine if socket closed
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			glog.Infof("%s: readFrom failed: %v", s.GetName(), err)
			continue
		}
		s.receiveCh <- buf[:n]
	}
}

func init() {
	transport.MustAddTransportType(typeName, NewUDP)
}
