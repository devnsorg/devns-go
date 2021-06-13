package mywgserver

import (
	"errors"
	"fmt"
	"github.com/ipTLS/ipTLS/wgapi/util"
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

type WGServer struct {
	endpoint *net.UDPAddr
	pool     *util.IPPool
	logger   *device.Logger
	errs     chan error
	iface    string
}

func NewWGServer(endpointAddressPort string, cidr string, logger *device.Logger, errs chan error) *WGServer {
	endpoint, err := net.ResolveUDPAddr("udp", endpointAddressPort)
	if err != nil {
		logger.Errorf("ERROR %#v", err)
		errs <- err
	}

	pool, err := util.NewIPPool(cidr)
	if err != nil {
		logger.Errorf("ERROR %#v", err)
		errs <- err
	}

	return &WGServer{endpoint: endpoint,
		pool: pool, logger: logger, errs: errs}
}

func (s *WGServer) StartServer() chan struct{} {
	isRoot, _ := checkIsRoot(s.logger)
	if !isRoot {
		s.errs <- errors.New("this program must be run as root! (sudo)")
	}

	iface, _, createdDevice := s.createDevice()
	s.iface = iface
	s.configureDevice()
	s.configureServerIP()

	return createdDevice.Wait()
}

func (s *WGServer) createDevice() (string, tun.Device, *device.Device) {
	interfaceName := "utun"
	createdTun, err := func() (tun.Device, error) {
		return tun.CreateTUN(interfaceName, device.DefaultMTU)
	}()

	if err != nil {
		s.logger.Errorf("CreateTUN error: %v", err)
		s.errs <- err
	}
	interfaceName, err = createdTun.Name()
	if err != nil {
		s.logger.Errorf("CreateTUN Name error: %v", err)
		s.errs <- err
	}

	fileUAPI, err := ipc.UAPIOpen(interfaceName)
	if err != nil {
		s.logger.Errorf("UAPI listen error: %v", err)
		s.errs <- err
	}

	createdDevice := device.NewDevice(createdTun, conn.NewDefaultBind(), s.logger)

	s.logger.Verbosef("Device started")

	uapi, err := ipc.UAPIListen(interfaceName, fileUAPI)
	if err != nil {
		s.logger.Errorf("Failed to listen on uapi socket: %v", err)
		s.errs <- err
	}

	go func() {
		for {
			createdConn, err := uapi.Accept()
			if err != nil {
				s.errs <- err
				return
			}
			go createdDevice.IpcHandle(createdConn)
		}
	}()

	s.logger.Verbosef("UAPI listener started")
	return interfaceName, createdTun, createdDevice
}

func (s *WGServer) configureDevice() {
	c, d := getUapi(s.iface, s.logger, s.errs)

	pk, err := wgtypes.GeneratePrivateKey()
	d.PrivateKey = pk

	listenPort := 51820

	err = c.ConfigureDevice(s.iface, wgtypes.Config{
		PrivateKey:   &pk,
		ListenPort:   &listenPort,
		FirewallMark: nil,
		ReplacePeers: true,
		Peers:        []wgtypes.PeerConfig{},
	})

	if err != nil {
		s.errs <- err
		s.logger.Errorf("ERROR %#v\n", err)
	}
}

func (s *WGServer) AddClientPeer() []byte {
	var err error
	c, d := getUapi(s.iface, s.logger, s.errs)

	serverIP := s.pool.GetStartingIP()
	clientIP, err := s.pool.Next(s.logger)
	if err != nil {
		s.logger.Errorf("pool.Next error: %v", err)
		s.errs <- err
	}

	duration, _ := time.ParseDuration("30s")
	peerKey, _ := wgtypes.GeneratePrivateKey()
	err = c.ConfigureDevice(s.iface, wgtypes.Config{
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   peerKey.PublicKey(),
				Remove:                      false,
				UpdateOnly:                  false,
				PersistentKeepaliveInterval: nil,
				ReplaceAllowedIPs:           true,
				AllowedIPs: []net.IPNet{{
					IP:   clientIP,
					Mask: s.pool.CurrentIPMask(),
				}},
			}},
	})
	if err != nil {
		s.errs <- err
		s.logger.Errorf("ConfigureDevice ERROR %#v\n", err)
	}

	wgQuickConfig := util.WgQuickConfig{

		Config: wgtypes.Config{
			PrivateKey:   &peerKey,
			ListenPort:   nil,
			ReplacePeers: true,
			Peers: []wgtypes.PeerConfig{
				{
					PublicKey:                   d.PublicKey,
					Remove:                      false,
					UpdateOnly:                  false,
					PresharedKey:                nil,
					Endpoint:                    s.endpoint,
					PersistentKeepaliveInterval: &duration,
					ReplaceAllowedIPs:           true,
					AllowedIPs: []net.IPNet{{
						IP:   serverIP,
						Mask: s.pool.CurrentIPMask(),
					}},
				},
			}},
		Address: []net.IPNet{{IP: clientIP,
			Mask: s.pool.CurrentIPMask()}},
	}
	configs, err := wgQuickConfig.MarshalText()
	s.logger.Verbosef("wgQuickConfig\n%s\n", configs)
	return configs
}

func (s *WGServer) configureServerIP() {
	serverIP := s.pool.GetStartingIP()
	serverIPNet := net.IPNet{
		IP:   serverIP,
		Mask: s.pool.CurrentIPMask(),
	}
	ifconfig := exec.Command("ifconfig", strings.Split(fmt.Sprintf("%s inet %s %s alias", s.iface, serverIPNet.String(), serverIP.String()), " ")...)
	stdoutStderr, err := ifconfig.CombinedOutput()
	if err != nil {
		s.logger.Errorf("ERROR %#v", err)
		s.errs <- err
	}
	s.logger.Verbosef("%s\n", stdoutStderr)

	//route := exec.Command("route", strings.Split("-q -n add -inet 10.44.0.1/24 -interface utun2", " ")...)
	//stdoutStderr, err = route.CombinedOutput()
	//if err != nil {
	//	logger.Errorf("ERROR %#v", err)
	//	errs <- err
	//}
	//logger.Verbosef("%s\n", stdoutStderr)
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

func getUapi(iface string, logger *device.Logger, errs chan error) (*wgctrl.Client, *wgtypes.Device) {
	uapiClient, err := wgctrl.New()
	if err != nil {
		logger.Errorf("wgctrl error: %v", err)
		errs <- err
	}
	devices, err := uapiClient.Devices()
	if err != nil {
		logger.Errorf("wgctrl get Devices error: %v", err)
		errs <- err
	}

	var uapiDevice *wgtypes.Device
	for _, iDevice := range devices {
		if iDevice.Name == iface {
			uapiDevice = iDevice
		}
	}

	if uapiDevice == nil {
		err = errors.New("device not found")
		errs <- err
		return nil, nil
	}
	return uapiClient, uapiDevice
}
