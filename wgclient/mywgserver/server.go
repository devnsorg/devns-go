package mywgserver

import (
	"errors"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func StartServer(logger *device.Logger, errsChan chan error) chan struct{} {
	isRoot, _ := checkIsRoot(logger)
	if !isRoot {
		errsChan <- errors.New("this program must be run as root! (sudo)")
	}

	createdDevice := createDevice(logger, errsChan)
	configureDevice(logger, errsChan)
	configureIPandRoute(logger, errsChan)
	return createdDevice.Wait()
}

func checkIsRoot(logger *device.Logger) (bool, error) {
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

func createDevice(logger *device.Logger, errs chan error) *device.Device {
	interfaceName := "utun"
	createdTun, err := func() (tun.Device, error) {
		return tun.CreateTUN(interfaceName, device.DefaultMTU)
	}()

	if err != nil {
		logger.Errorf("CreateTUN error: %v", err)
		errs <- err
	}
	interfaceName, err = createdTun.Name()
	if err != nil {
		logger.Errorf("CreateTUN Name error: %v", err)
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

func configureDevice(logger *device.Logger, errs chan error) {
	c, err := wgctrl.New()
	if err != nil {
		logger.Errorf("wgctrl error: %v", err)
		errs <- err
	}
	ds, err := c.Devices()
	if err != nil {
		logger.Errorf("wgctrl get Devices error: %v", err)
		errs <- err
	}
	iface := ds[0].Name
	d := ds[0]
	pk, err := wgtypes.ParseKey("CCWeMw4sl6USwstoCIFKf7pnivn5bG94eyALNiEHyuM=")
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

func configureIPandRoute(logger *device.Logger, errs chan error) {
	//TODO Add inet and route

	ifconfig := exec.Command("ifconfig", strings.Split("utun2 inet 10.44.0.17/32 10.44.0.17 alias", " ")...)
	stdoutStderr, err := ifconfig.CombinedOutput()
	if err != nil {
		logger.Errorf("ERROR %#v", err)
		errs <- err
	}
	logger.Verbosef("%s\n", stdoutStderr)

	route := exec.Command("route", strings.Split("-q -n add -inet 10.44.0.1/24 -interface utun2", " ")...)
	stdoutStderr, err = route.CombinedOutput()
	if err != nil {
		logger.Errorf("ERROR %#v", err)
		errs <- err
	}
	logger.Verbosef("%s\n", stdoutStderr)
}
