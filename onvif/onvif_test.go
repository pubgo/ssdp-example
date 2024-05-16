package onvif_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/use-go/onvif"
)

func TestHNet(t *testing.T) {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(fmt.Errorf("localAddresses: %v", err.Error()))
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Println(fmt.Errorf("localAddresses: %v", err.Error()))
			continue
		}

		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				fmt.Printf("%v : %s [%v/%v]\n", i.Name, v, v.IP, v.Mask)
			default:
				panic(v)
			}
		}
	}
}

func TestGetAvailableDevicesAtSpecificEthernetInterface(t *testing.T) {
	s, err := onvif.GetAvailableDevicesAtSpecificEthernetInterface("en0")
	if err != nil {
		panic(err)
	}

	for i := range s {
		fmt.Printf("GetDeviceInfo %#v\n", s[i].GetDeviceInfo())
		for k, v := range s[i].GetServices() {
			fmt.Println(k, v)
		}
	}
}
