package main

import (
	"flag"
	"fmt"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var portF = flag.Int("port", 8888, "Port for DNS server to listen to")
var rootDomainF = flag.String("domain", "example.com", "[MUST CHANGE] Base domain for DNS resolution")
var helpF = flag.Bool("h", false, "Print this help")

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	flag.Parse()

	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	i, err := strconv.Atoi(string(output[:len(output)-1]))

	if err != nil {
		log.Fatal(err)
	}

	if i > 0 {
		log.Fatal("This program must be run as root! (sudo)")
	}

	interfaceName := "utun"
	logger := device.NewLogger(
		device.LogLevelVerbose,
		fmt.Sprintf("(%s) ", interfaceName),
	)

	createdTun, err := func() (tun.Device, error) {
		return tun.CreateTUN(interfaceName, device.DefaultMTU)
	}()

	if err == nil {
		realInterfaceName, err2 := createdTun.Name()
		if err2 == nil {
			interfaceName = realInterfaceName
		}
	} else {
		log.Fatalf("Failed due to %#v\n", err)
	}

	fileUAPI, err := ipc.UAPIOpen(interfaceName)
	if err != nil {
		logger.Errorf("UAPI listen error: %v", err)
		os.Exit(-1)
		return
	}

	createdDevice := device.NewDevice(createdTun, conn.NewDefaultBind(), logger)
	defer createdDevice.Close()

	logger.Verbosef("Device started")

	errs := make(chan error)
	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM)
	signal.Notify(term, os.Interrupt)

	uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	if err != nil {
		logger.Errorf("Failed to listen on uapi socket: %v", err)
		os.Exit(-1)
	}

	go func() {
		for {
			createdConn, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go createdDevice.IpcHandle(createdConn)
		}
	}()

	logger.Verbosef("UAPI listener started")

	c, _ := wgctrl.New()
	ds, _ := c.Devices()
	iface := ds[0].Name
	logger.Verbosef("DEvices %#v %s", ds, iface)
	d := ds[0]
	pk, _ := wgtypes.ParseKey("CCWeMw4sl6USwstoCIFKf7pnivn5bG94eyALNiEHyuM=")
	d.PrivateKey = pk

	pubk, _ := wgtypes.ParseKey("k50wd3iSagK+vlZDh8KAEYYitaYeUrxdf+fM8o83cRg=")

	endpoint, _ := net.ResolveUDPAddr("udp", "vpn.tripath.vn:51820")
	_, ipNet, err := net.ParseCIDR("10.44.0.1/32")
	duration, _ := time.ParseDuration("30s")
	err = c.ConfigureDevice(iface, wgtypes.Config{
		PrivateKey:   &pk,
		ListenPort:   nil,
		FirewallMark: nil,
		ReplacePeers: true,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   pubk,
				Remove:                      false,
				UpdateOnly:                  false,
				PresharedKey:                nil,
				Endpoint:                    endpoint,
				PersistentKeepaliveInterval: &duration,
				ReplaceAllowedIPs:           true,
				AllowedIPs:                  []net.IPNet{*ipNet},
			},
		}})
	if err != nil {
		log.Fatalf("ERROR %#v\n", err)
	}
	//currentDevice, err := c.Device(iface)
	//if err != nil {
	//	log.Fatalf("ERROR %#v\n", err)
	//}
	//
	//_, deviceInet, _ := net.ParseCIDR("10.44.0.17/32")

	// FOR DARWIN
	/*
		route -q -n add -inet 10.44.0.1/32 -interface utun2
		sudo route -q -n add -inet 10.44.0.1/32 -interface utun2
	*/

	select {
	case <-term:
	case <-errs:
	case <-createdDevice.Wait():
	}
}
