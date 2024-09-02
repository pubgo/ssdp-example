package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/beevik/etree"
	"github.com/emiago/sipgo"
	sip2 "github.com/emiago/sipgo/sip"
	"github.com/ghettovoice/gosip"
	"github.com/ghettovoice/gosip/log"
	"github.com/ghettovoice/gosip/sip"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	discover "github.com/use-go/onvif/ws-discovery"

	_ "github.com/use-go/onvif/sdk/device"
	sdk "github.com/use-go/onvif/sdk/device"

	_ "github.com/emiago/sipgo"
	_ "github.com/ghettovoice/gosip"
)

var name = flag.String("name", "en0", "Ethernet Interface Name")

func main() {
	flag.Parse()

	//server()
	//server1()
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
			}
		}
	}

	s, err := onvif.GetAvailableDevicesAtSpecificEthernetInterface(*name)
	if err != nil {
		panic(err)
	}

	for i := range s {
		fmt.Println(s[i].GetDeviceInfo())
		fmt.Printf("GetDeviceInfo %#v\n", s[i].GetDeviceInfo())
		for k, v := range s[i].GetServices() {
			if strings.Contains(v, "172.20.22.146") {
				fmt.Println(k, v)
			}
		}
	}

	//client1()
}

func client1() {
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr: "172.20.22.146",
		// Username: "admin",
		// Password: "tm1234",
	})
	if err != nil {
		panic(err)
	}

	// fmt.Printf("output %+v", dev.GetServices())

	// fmt.Printf("%#v\n\n", dev.GetDeviceInfo())
	// fmt.Printf("%#v\n\n", dev.GetServices())
	rsp := must(sdk.Call_GetDeviceInformation(context.Background(), dev, device.GetDeviceInformation{}))
	fmt.Printf("output %#v\n", rsp)

	fmt.Printf("%#v\n", must(sdk.Call_GetSystemSupportInformation(context.Background(), dev, device.GetSystemSupportInformation{})))

	rsp1 := must(sdk.Call_GetCapabilities(context.Background(), dev, device.GetCapabilities{Category: "All"}))
	fmt.Printf("output %#v\n", rsp1)
}

func must[T any](d T, err error) T {
	if err != nil {
		panic(err)
	}
	return d
}

// Host host
type Host struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

func runDiscovery(interfaceName string) {
	var hosts []*Host
	devices, err := discover.SendProbe(interfaceName, nil, []string{"dn:NetworkVideoTransmitter"}, map[string]string{"dn": "http://www.onvif.org/ver10/network/wsdl"})
	if err != nil {
		fmt.Printf("error %s", err)
		return
	}
	for _, j := range devices {
		doc := etree.NewDocument()
		if err := doc.ReadFromString(j); err != nil {
			fmt.Printf("error %s", err)
		} else {

			endpoints := doc.Root().FindElements("./Body/ProbeMatches/ProbeMatch/XAddrs")
			scopes := doc.Root().FindElements("./Body/ProbeMatches/ProbeMatch/Scopes")

			flag := false

			host := &Host{}

			for _, xaddr := range endpoints {
				xaddr := strings.Split(strings.Split(xaddr.Text(), " ")[0], "/")[2]
				host.URL = xaddr
			}
			if flag {
				break
			}
			for _, scope := range scopes {
				re := regexp.MustCompile(`onvif:\/\/www\.onvif\.org\/name\/[A-Za-z0-9-]+`)
				match := re.FindStringSubmatch(scope.Text())
				host.Name = path.Base(match[0])
			}

			hosts = append(hosts, host)

		}

	}

	bys, _ := json.Marshal(hosts)
	fmt.Printf("done %s", bys)
}

func server() {
	srvConf := gosip.ServerConfig{}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	srv := gosip.NewServer(srvConf, nil, nil, log.NewDefaultLogrusLogger())
	must(1, srv.OnRequest(sip.INVITE, func(req sip.Request, tx sip.ServerTransaction) {
	}))
	srv.Listen("udp", "0.0.0.0:5081")
	<-stop
	srv.Shutdown()
}

func server1() {
	ua, err := sipgo.NewUA()
	if err != nil {
		panic(err)
	}

	go func() {
		time.Sleep(time.Second * 10)

		client, err := sipgo.NewClient(ua, sipgo.WithClientHostname("172.20.22.76"))
		if err != nil {
			panic(err)
		}
		defer client.Close()

		recipient := &sip2.Uri{}
		assert(sip2.ParseUri(fmt.Sprintf("sip:%s@%s", "client:123456", "172.20.22.76:5081"), recipient))

		req := sip2.NewRequest(sip2.REGISTER, *recipient)
		req.AppendHeader(sip2.NewHeader("Contact", fmt.Sprintf("<sip:%s@%s>", "client", "172.20.22.76")))
		req.SetTransport(strings.ToUpper("udp"))
		tx, err := client.TransactionRequest(context.Background(), req)
		if err != nil {
			panic(err)
		}
		defer tx.Terminate()
		res, err := getResponse(tx)
		if err != nil {
			panic(err)
		}
		fmt.Println(res.String())

		for {
			recipient1 := &sip2.Uri{}
			assert(sip2.ParseUri(fmt.Sprintf("sip:%s@%s", "123456:123456", "172.20.22.146:80"), recipient1))
			assert(client.WriteRequest(sip2.NewRequest(sip2.INVITE, *recipient1)))
			time.Sleep(time.Second * 5)
		}
	}()

	srv := must(sipgo.NewServer(ua))
	srv.OnSubscribe(func(req *sip2.Request, tx sip2.ServerTransaction) {
		//req.GetHeaders()
	})

	srv.OnRegister(func(req *sip2.Request, tx sip2.ServerTransaction) {
		for _, h := range req.Headers() {
			fmt.Println("header: ", h.String())
		}

		fmt.Println("method: ", req.Method.String())
		fmt.Println("ver: ", req.SipVersion)
		fmt.Println("to: ", req.To().String())

		fmt.Println(tx.Respond(sip2.NewResponseFromRequest(req, 200, "OK", nil)))
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stop
		cancel()
	}()

	srv.OnBye(func(req *sip2.Request, tx sip2.ServerTransaction) {
		assert(tx.Respond(sip2.NewResponseFromRequest(req, 200, "OK", nil)))
	})

	srv.OnAck(func(req *sip2.Request, tx sip2.ServerTransaction) {
		assert(tx.Respond(sip2.NewResponseFromRequest(req, 200, "OK", nil)))
	})

	srv.ListenAndServe(ctx, "udp", "0.0.0.0:5081")

	<-ctx.Done()
	srv.Close()
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func getResponse(tx sip2.ClientTransaction) (*sip2.Response, error) {
	select {
	case <-tx.Done():
		return nil, fmt.Errorf("transaction died")
	case res := <-tx.Responses():
		return res, nil
	}
}
