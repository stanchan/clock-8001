package clock

import (
	"fmt"
	"gitlab.com/Depili/clock-8001/v3/debug"
	"log"
	"net"
	"strings"
	"time"
)

const interfacePollTime = 5 * time.Second

type feedbackDestination struct {
	udpConns []*net.UDPConn
	address  string
}

func initFeedback(address string) *feedbackDestination {
	var fbDest = feedbackDestination{
		address: address,
	}
	go fbDest.monitor()
	return &fbDest
}

func (fbDest *feedbackDestination) Write(data []byte) {
	debug.Printf("Writing data to connections\n")
	for _, conn := range fbDest.udpConns {
		if _, err := conn.Write(data); err != nil {
			debug.Printf(" -> Error writing to udp connection %v", conn)
		}
	}
}

func (fbDest *feedbackDestination) monitor() {
	log.Printf("Monitoring network interface changes\n")
	port := strings.Join(strings.Split(fbDest.address, ":")[1:], "")
	log.Printf("Feedback port: %v", port)

	for {
		time.Sleep(interfacePollTime)
		log.Printf("Updating feedback connections\n")

		if !strings.Contains(fbDest.address, "255.255.255.255") {
			fbDest.singleAddr()
			continue
		}
		fbDest.broadcastAll(port)
	}
}

func (fbDest *feedbackDestination) singleAddr() {
	log.Printf(" -> Trying single address: %v\n", fbDest.address)
	udpConns := make([]*net.UDPConn, 0)

	if udpAddr, err := net.ResolveUDPAddr("udp", fbDest.address); err != nil {
		log.Printf(" -> Failed to resolve feedback address: %v", err)
	} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
		log.Printf("   -> Failed to open feedback address: %v", err)
	} else {
		log.Printf("Feedback: sending to %v", fbDest.address)
		udpConns = append(udpConns, udpConn)
	}
	fbDest.udpConns = udpConns
}

func (fbDest *feedbackDestination) broadcastAll(port string) {
	log.Printf(" -> Broadcasting to all interfaces\n")
	udpConns := make([]*net.UDPConn, 0)

	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		ip, n, err := net.ParseCIDR(addr.String())
		if err != nil {
			log.Printf(" -> error parsing network: %v\n", err)
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
					log.Printf(" -> Failed to resolve broadcast address %v: %v", dest, err)
				} else if udpConn, err := net.DialUDP("udp", nil, udpAddr); err != nil {
					log.Printf("   -> Failed to open broadcast address %v: %v", dest, err)
				} else {
					log.Printf("Feedback: sending to %v", dest)
					udpConns = append(udpConns, udpConn)
				}
			}
		}
	}
	fbDest.udpConns = udpConns
}

func (fbDest *feedbackDestination) String() string {
	return fbDest.address
}
