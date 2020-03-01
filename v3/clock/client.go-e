package clock

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
)

// ClientOptions common options for client instances
type ClientOptions struct {
	Connect string `long:"clock-client-connect" description:"Address to send clock osc messages to" default:"255.255.255.255:1245"` // Address to connect to with OSC
}

// MakeClient Create a clock OSC client
func (options ClientOptions) MakeClient() (*Client, error) {
	var client = Client{}
	client.oscDests = initFeedback(options.Connect)

	return &client, nil
}

// Client A clock osc client
type Client struct {
	oscDests *feedbackDestination
}

// Print the connection info of a Client
func (client *Client) String() string {
	return fmt.Sprintf("%v", client.oscDests.String())
}

func (client *Client) send(packet osc.Packet) error {
	if client.oscDests == nil {
		// No osc connection
		return nil
	}

	data, err := packet.MarshalBinary()
	if err != nil {
		return err
	}

	client.oscDests.Write(data)
	return nil
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
