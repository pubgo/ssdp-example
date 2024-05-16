package main

// https://github.com/hashicorp/mdns
// https://github.com/pion/mdns
// https://github.com/grandcat/zeroconf
// https://github.com/arduino/mdns-discovery
// https://github.com/blake/external-mdns
// https://github.com/brutella/dnssd
// https://github.com/RedTeamPentesting/pretender
// https://github.com/openshift/coredns-mdns
//
import (
	"fmt"
	"time"

	"github.com/hashicorp/mdns"
	_ "github.com/hashicorp/mdns"
)

func main() {
	// go func() {
	// 	host, _ := os.Hostname()
	// 	info := []string{"My awesome service"}
	// 	service, _ := mdns.NewMDNSService(host, "_foobar._tcp", "", "", 8000, nil, info)

	// 	// Create the mDNS server, defer shutdown
	// 	server, _ := mdns.NewServer(&mdns.Config{Zone: service})
	// 	defer server.Shutdown()

	// 	<-time.After(time.Second * 10)
	// }()
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %#v\n", *entry)
		}
	}()

	// Start the lookup
	mdns.Lookup("*._tcp", entriesCh)
	close(entriesCh)
	<-time.After(time.Second * 10)
}
