package clock

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"log"
	"net"
	"strings"
	"time"
)

// ClientOptions common options for client instances
type ClientOptions struct {
	Connect string `long:"clock-client-connect" description:"Address to send clock osc messages to" default:"255.255.255.255:1245"` // Address to connect to with OSC
}

// MakeClient Create a clock OSC client
func (options ClientOptions) MakeClient() (*Client, error) {
	var client = Client{}

	client.udpDest = options.Connect
	// Poll for network interface changes
	go client.interfaceMonitor()

	return &client, nil
}

// Client A clock osc client
type Client struct {
	udpDest  string
	oscDests *feedbackDestinations
}

// Monitor for interface address changes and update broadcast destinations
func (client *Client) interfaceMonitor() {
	log.Printf("Monitoring network interface changes\n")
	port := strings.Join(strings.Split(client.udpDest, ":")[1:], "")
	log.Printf("OSC feedback port: %v", port)

	for {
		time.Sleep(interfacePollTime)
		log.Printf("Updating feedback connections\n")

		conns := feedbackDestinations{
			udpConns: make([]*net.UDPConn, 0),
		}

		if !strings.Contains(client.udpDest, "255.255.255.255") {
			log.Printf(" -> Trying single address: %v\n", client.udpDest)
			if udpAddr, err := net.ResolveUDPAddr("udp", client.udpDest); err != nil {
				log.Printf(" -> Failed to resolve OSC feedback address: %v", err)
			} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
				log.Printf("   -> Failed to open OSC feedback address: %v", err)
			} else {
				log.Printf("OSC feedback: sending to %v", client.udpDest)
				conns.udpConns = append(conns.udpConns, udpConn)
			}
			continue
		}

		addrs, _ := net.InterfaceAddrs()
		for _, addr := range addrs {
			ip, n, err := net.ParseCIDR(addr.String())
			if err != nil {
				log.Printf(" -> error parsing network\n")
			} else {
				if ip.IsLoopback() {
					// Ignore loopback interfaces
					continue
				} else if ip.To4() != nil {
					broadcast := net.IP(make([]byte, 4))
					for i := range n.IP {
						broadcast[i] = n.IP[i] | (^n.Mask[i])
					}
					log.Printf(" -> using broadcast address %v", broadcast)

					dest := fmt.Sprintf("%v:%v", broadcast, port)

					if udpAddr, err := net.ResolveUDPAddr("udp", dest); err != nil {
						log.Printf(" -> Failed to resolve OSC broadcast address %v: %v", dest, err)
					} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
						log.Printf("   -> Failed to open OSC broadcast address %v: %v", dest, err)
					} else {
						log.Printf("OSC feedback: sending to %v", dest)
						conns.udpConns = append(conns.udpConns, udpConn)
					}
				}
			}
		}
		client.oscDests = &conns
	}
}

// Print the connection info of a Client
func (client *Client) String() string {
	return fmt.Sprintf("%v", client.udpDest)
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

	for _, conn := range client.oscDests.udpConns {
		if _, err := conn.Write(data); err != nil {
			return err
		}
	}
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
