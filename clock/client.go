package clock

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"net"
)

type ClientOptions struct {
	Connect string `long:"clock-client-connect"`
}

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

type Client struct {
	udpConn *net.UDPConn
}

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

func (client *Client) SendCount(message CountMessage) error {
	return client.send(message.MarshalOSC("/qmsk/clock/count"))
}

func (client *Client) SendStart(message StartMessage) error {
	return client.send(message.MarshalOSC("/clock/countdown/start"))
}
