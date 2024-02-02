package UDPreceiver

import (
	"encoding/gob"
	"fmt"
	"net"
	"bytes"
	"deltadiode/Message"
)

// UDPreceiver represents a UDP receiver that listens for messages and sends them to a channel.
type UDPreceiver struct {
	address       string
	interfaceName string
	conn          *net.UDPConn
	MessageCh     chan Message.Message // Channel for sending received messages
}

// NewUDPreceiver creates a new UDPreceiver with the given address.
func NewUDPreceiver(address, interfaceName string) *UDPreceiver {
	return &UDPreceiver{
		address:       address,
		interfaceName: interfaceName,
		MessageCh:     make(chan Message.Message),
	}
}

// Start initiates the UDP receiver to listen for incoming messages.
func (r *UDPreceiver) Start() error {
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
//	fmt.Println("UDP receiver started on", r.address)

	go func() {
		defer r.conn.Close()
		buffer := make([]byte, 10240)

		for {
			n, _, err := r.conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error receiving data:", err)
				continue
			}

			var networkData bytes.Buffer
			networkData.Write(buffer[:n])
			dec := gob.NewDecoder(&networkData)

			var packet Message.Message
			err = dec.Decode(&packet)
			if err != nil {
				fmt.Println("decode error:", err)
				continue
			}

			// Send the received message to the channel
			r.MessageCh <- packet
		}
	}()

	return nil
}
