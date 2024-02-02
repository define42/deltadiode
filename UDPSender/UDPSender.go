package UDPSender

import (
	"bytes"
	"deltadiode/Message"
	"encoding/gob"
	"fmt"
	"net"
)

// UDPSender represents a UDP sender that sends messages read from a channel.
type UDPSender struct {
	address       string
	interfaceName string
	messages      chan []byte
	conn          *net.UDPConn
}

// NewUDPSender creates a new UDPSender with the given address.
func NewUDPSender(address, interfaceName string) *UDPSender {
	return &UDPSender{
		address:       address,
		interfaceName: interfaceName,
		messages:      make(chan []byte),
	}
}

// Start initiates the process of listening to the messages channel and sending messages over UDP.
func (s *UDPSender) Start() error {

	groupAddr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		fmt.Println("Error resolving group address:", err)
		return err
	}

	iface, err := net.InterfaceByName(s.interfaceName)
	if err != nil {
		fmt.Println("Error finding network interface:", err)
		return err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		fmt.Println("Error getting interface addresses:", err)
		return err
	}

	var localIP net.IP
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				localIP = ipnet.IP
			}
		}
	}

	if localIP == nil {
		fmt.Println("No suitable IPv4 address found on the interface")
		return err
	}

	localAddr := &net.UDPAddr{
		IP:   localIP,
		Port: 0,
	}

	conn, err := net.DialUDP("udp", localAddr, groupAddr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		return err
	}

	s.conn = conn

	go func() {
		for msg := range s.messages {
			_, err := s.conn.Write([]byte(msg))
			if err != nil {
				fmt.Println("Error sending message:", err)
			}
		}
		s.conn.Close()
	}()

	return nil
}

// Send queues a message to be sent over UDP.
func (s *UDPSender) SendHash(msg []byte) {
	s.messages <- msg
}

// Send queues a message to be sent over UDP.
func (s *UDPSender) Send(msg Message.Message) error {

	var buf bytes.Buffer
	gobenc := gob.NewEncoder(&buf)
	err := gobenc.Encode(msg)
	if err != nil {
		return err
	}

	s.messages <- buf.Bytes()
	return nil
}

// Close closes the messages channel and the UDP connection.
func (s *UDPSender) Close() {
	close(s.messages)
}
