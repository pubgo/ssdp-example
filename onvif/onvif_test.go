package onvif_test

import (
	"fmt"
	"net"
	"testing"
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
