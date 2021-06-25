package mywgserver

import (
	"errors"
	"github.com/devnsorg/devns-go/util"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"os/exec"
	"strconv"
	"time"
)

type WGServer struct {
	endpoint      *net.UDPAddr
	ipPool        *util.IPPool
	logger        *device.Logger
	errs          chan error
	iface         string
	createdTun    tun.Device
	createdDevice *device.Device
	uapiListen    net.Listener
	duration      time.Duration
	wgPool        *util.WGPool
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
	duration, _ := time.ParseDuration("5s")
	return &WGServer{
		endpoint: endpoint,
		ipPool:   pool,
		logger:   logger,
		errs:     errs,
		duration: duration,
	}
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
	s.wgPool = util.NewWGPool(iface, s.logger, s.errs)
	go s.wgPool.CleanUpStalePeers(s.duration, func(poolPeer *util.WGPoolPeer) {
		c, _ := util.GetUapi(s.iface, s.logger, s.errs)
		err := c.ConfigureDevice(iface, wgtypes.Config{
			ReplacePeers: false,
			Peers: []wgtypes.PeerConfig{{
				PublicKey:  poolPeer.PublicKey(),
				Remove:     true,
				UpdateOnly: true,
			}},
		})
		if err != nil {
			s.logger.Errorf("CleanUpStalePeers %#v %#v", poolPeer, err)
			s.errs <- err
		}

	})

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

	uapiListen, err := ipc.UAPIListen(interfaceName, fileUAPI)
	if err != nil {
		s.logger.Errorf("Failed to listen on uapiListen socket: %v", err)
		s.errs <- err
	}

	go func() {
		for {
			createdConn, err := uapiListen.Accept()
			if err != nil {
				s.errs <- err
				return
			}
			go createdDevice.IpcHandle(createdConn)
		}
	}()

	s.logger.Verbosef("UAPI listener started")
	s.createdTun = createdTun
	s.createdDevice = createdDevice
	s.uapiListen = uapiListen
	return interfaceName, createdTun, createdDevice
}

func (s *WGServer) configureDevice() {
	c, _ := util.GetUapi(s.iface, s.logger, s.errs)

	pk, err := wgtypes.GeneratePrivateKey()

	listenPort := s.endpoint.Port

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

func (s *WGServer) AddClientPeer(subdomain string) []byte {
	var err error
	c, d := util.GetUapi(s.iface, s.logger, s.errs)

	serverIP := s.ipPool.GetStartingIP()
	clientIP, err := s.ipPool.Next(s.logger)
	if err != nil {
		s.logger.Errorf("ipPool.Next error: %v", err)
		s.errs <- err
	}

	peerKey, _ := wgtypes.GeneratePrivateKey()
	err = c.ConfigureDevice(s.iface, wgtypes.Config{
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   peerKey.PublicKey(),
				PersistentKeepaliveInterval: &s.duration,
				AllowedIPs: []net.IPNet{{
					IP:   clientIP,
					Mask: net.CIDRMask(32, 32),
				}},
			}},
	})
	if err != nil {
		s.errs <- err
		s.logger.Errorf("ConfigureDevice ERROR %#v\n", err)
	}

	wgQuickConfig := util.WgQuickConfig{

		Config: wgtypes.Config{
			PrivateKey: &peerKey,
			Peers: []wgtypes.PeerConfig{
				{
					PublicKey:                   d.PublicKey,
					Endpoint:                    s.endpoint,
					PersistentKeepaliveInterval: &s.duration,
					AllowedIPs: []net.IPNet{{
						IP:   serverIP,
						Mask: s.ipPool.CurrentIPMask(),
					}},
				},
			}},
		Address: []net.IPNet{{
			IP:   clientIP,
			Mask: s.ipPool.CurrentIPMask(),
		}},
	}

	configs, err := wgQuickConfig.MarshalText()
	s.logger.Verbosef("wgQuickConfig\n%s\n", configs)
	configString, _ := wgQuickConfig.MarshalText()

	s.wgPool.AddPoolPeer(subdomain, peerKey.PublicKey(), clientIP)

	return configString
}

func (s *WGServer) StopServer() {
	_ = s.uapiListen.Close()
	_ = s.createdTun.Close()
	s.createdDevice.Close()
}

func (s *WGServer) GetPeerAddressFor(subdomain string) net.IP {
	s.logger.Verbosef("GetPeerAddressFor %s", subdomain)
	if s.wgPool.HasSubdomain(subdomain) {
		return s.wgPool.GetPeerAddressBySubdomain(subdomain)
	} else {
		return nil
	}
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
