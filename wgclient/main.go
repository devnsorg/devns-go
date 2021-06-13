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
	"strings"
	"syscall"
	"time"
)

var portF = flag.Int("port", 8888, "Port for DNS server to listen to")
var rootDomainF = flag.String("domain", "example.com", "[MUST CHANGE] Base domain for DNS resolution")
var helpF = flag.Bool("h", false, "Print this help")

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	flag.Parse()

	logger := device.NewLogger(
		device.LogLevelVerbose,
		fmt.Sprintf("(WG DEVICE) "),
	)

	isRoot, _ := CheckIsRoot(logger)
	if !isRoot {
		logger.Errorf("This program must be run as root! (sudo)")
		return
	}

	errsChan := make(chan error)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		select {
		case <-signalChan:
			os.Exit(0)
		case err := <-errsChan:
			logger.Errorf("ERRSCHAN %#v", err)
			os.Exit(2)
		}
	}()

	createdDevice := CreateDevice(logger, errsChan)
	ConfigureDevice(logger, errsChan)
	ConfigureIPandRoute(logger, errsChan)

	select {
	case <-createdDevice.Wait():
	}
}

func ConfigureIPandRoute(logger *device.Logger, errs chan error) {
	//TODO Add inet and route

	ifconfig := exec.Command("ifconfig", strings.Split("utun2 inet 10.44.0.17/32 10.44.0.17 alias", " ")...)
	stdoutStderr, err := ifconfig.CombinedOutput()
	if err != nil {
		logger.Errorf("ERROR %#v", err)
		errs <- err
	}
	fmt.Printf("%s\n", stdoutStderr)

	route := exec.Command("route", strings.Split("-q -n add -inet 10.44.0.1/24 -interface utun2", " ")...)
	stdoutStderr, err = route.CombinedOutput()
	if err != nil {
		logger.Errorf("ERROR %#v", err)
		errs <- err
	}
	fmt.Printf("%s\n", stdoutStderr)
}

func CheckIsRoot(logger *device.Logger) (bool, error) {
	var err error
	cmd := exec.Command("id", "-u")
	output, err := cmd.Output()

	if err != nil {
		logger.Errorf("ERROR %#v", err)
	}

	i, err := strconv.Atoi(string(output[:len(output)-1]))

	if err != nil {
		logger.Errorf("ERROR %#v", err)
	}

	if i > 0 {
		return false, err
	} else {
		return true, err
	}
}

func CreateDevice(logger *device.Logger, errs chan error) *device.Device {
	interfaceName := "utun"
	createdTun, err := func() (tun.Device, error) {
		return tun.CreateTUN(interfaceName, device.DefaultMTU)
	}()

	if err == nil {
		realInterfaceName, err2 := createdTun.Name()
		if err2 == nil {
			interfaceName = realInterfaceName
		}
	} else {
		errs <- err
	}

	fileUAPI, err := ipc.UAPIOpen(interfaceName)
	if err != nil {
		logger.Errorf("UAPI listen error: %v", err)
		errs <- err
	}

	createdDevice := device.NewDevice(createdTun, conn.NewDefaultBind(), logger)

	logger.Verbosef("Device started")

	uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	if err != nil {
		logger.Errorf("Failed to listen on uapi socket: %v", err)
		errs <- err
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
	return createdDevice
}

func ConfigureDevice(logger *device.Logger, errs chan error) {
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
	if err != nil {
		errs <- err
		logger.Errorf("ERROR %#v\n", err)
	}
	allowedIPs := []net.IPNet{*ipNet}
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
				AllowedIPs:                  allowedIPs,
			},
		}})
	if err != nil {
		errs <- err
		logger.Errorf("ERROR %#v\n", err)
	}
}
