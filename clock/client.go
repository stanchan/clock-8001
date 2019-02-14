package clock

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"net"
)

// ClientOptions common options for client instances
type ClientOptions struct {
	Connect string `long:"clock-client-connect"` // Address to connect to with OSC
}

// MakeClient Create a clock OSC client
func (options ClientOptions) MakeClient() (*Client, error) {
	var client = Client{}

	if udpAddr, err := net.ResolveUDPAddr("udp", options.Connect); err != nil {
		return nil, err
	} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
		return nil, err
	} else {
		client.udpConn = udpConn
	}

	return &client, nil
}

// Client A clock osc client
type Client struct {
	udpConn *net.UDPConn
}

// Print the connection info of a Client
func (client *Client) String() string {
	return fmt.Sprintf("%v", client.udpConn.RemoteAddr())
}

func (client *Client) send(packet osc.Packet) error {
	if data, err := packet.MarshalBinary(); err != nil {
		return err
	} else if _, err := client.udpConn.Write(data); err != nil {
		return err
	} else {
		return nil
	}
}

// SendDisplay Send a /clock/display message
func (client *Client) SendDisplay(message DisplayMessage) error {
	return client.send(message.MarshalOSC("/clock/display"))
}

// SendCount Send a /clock/count message
func (client *Client) SendCount(message CountMessage) error {
	return client.send(message.MarshalOSC("/qmsk/clock/count"))
}

// SendStart Send a /clock/countdown/start message
func (client *Client) SendStart(message CountdownMessage) error {
	return client.send(message.MarshalOSC("/clock/countdown/start"))
}
