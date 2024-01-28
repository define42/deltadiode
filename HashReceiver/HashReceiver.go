package HashReceiver

import (
	"fmt"
	"net"
)

type HASHreceiver struct {
	address       string
	interfaceName string
	conn          *net.UDPConn
	MessageCh     chan []byte
}

func NewHASHreceiver(address, interfaceName string) *HASHreceiver {
	return &HASHreceiver{
		address:       address,
		interfaceName: interfaceName,
		MessageCh:     make(chan []byte),
	}
}

func (r *HASHreceiver) Start() error {
	// Resolve the UDP address for multicast
	udpAddr, err := net.ResolveUDPAddr("udp", r.address)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Get the network interface by name
	iface, err := net.InterfaceByName(r.interfaceName)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Create a UDP connection using the specified network interface
	conn, err := net.ListenMulticastUDP("udp", iface, udpAddr)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	r.conn = conn

	go func() {
		defer r.conn.Close()
		buffer := make([]byte, 10240)

		for {
			n, _, err := r.conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error receiving data:", err)
				continue
			}
			if n != 32 {
				fmt.Println("Error: This is not SHA256 32byte")
				continue
			}

			r.MessageCh <- buffer[:n]
		}
	}()

	return nil
}
