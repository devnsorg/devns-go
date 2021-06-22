package mywgserver

import (
	"errors"
	"github.com/devnsorg/devns-go/wgapi/util"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"os/exec"
	"strconv"
	"time"
)

type WGServer struct {
	endpoint      *net.UDPAddr
	pool          *util.IPPool
	logger        *device.Logger
	errs          chan error
	iface         string
	createdTun    tun.Device
	createdDevice *device.Device
	uapiListen    net.Listener
	duration      time.Duration
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
	duration, _ := time.ParseDuration("30s")
	return &WGServer{endpoint: endpoint,
		pool: pool, logger: logger, errs: errs, duration: duration}
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
	go s.cleanUpStalePeers()
	return createdDevice.Wait()
}

var subdomains = make(map[string]*util.WgQuickConfig)
var subdomainsZeroHandshake = make(map[string]bool)

func (s *WGServer) GetPeerAddressFor(subdomain string) net.IP {
	return subdomains[subdomain].Address[0].IP
}

func (s *WGServer) cleanUpStalePeers() {
	c, d := getUapi(s.iface, s.logger, s.errs)
	for range time.Tick(time.Second * 1) {
		for _, peer := range d.Peers {
			var subdomain = ""
			for iSubdomain, config := range subdomains {
				if config.PrivateKey.PublicKey() == peer.PublicKey {
					s.logger.Verbosef("REMOVE PEER Map %s", iSubdomain)
					subdomain = iSubdomain
				}
			}
			if len(subdomain) == 0 {
				s.errs <- errors.New("Map not synced")
			}

			if peer.LastHandshakeTime.IsZero() && !subdomainsZeroHandshake[subdomain] {
				subdomainsZeroHandshake[subdomain] = true
			} else if peer.LastHandshakeTime.Add(2 * s.duration).Before(time.Now()) {
				// If 2xDURATION passed, delete peer
				err := c.ConfigureDevice(s.iface, wgtypes.Config{
					ReplacePeers: false,
					Peers: []wgtypes.PeerConfig{{
						PublicKey:  peer.PublicKey,
						Remove:     true,
						UpdateOnly: true,
					}},
				})
				if err != nil {
					s.logger.Errorf("REMOVE PEER ConfigureDevice %#v", err)
					s.errs <- err
				}
				delete(subdomains, subdomain)
				delete(subdomainsZeroHandshake, subdomain)
			}
		}
	}
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
	c, _ := getUapi(s.iface, s.logger, s.errs)

	pk, err := wgtypes.GeneratePrivateKey()

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

func (s *WGServer) AddClientPeer(subdomain string) []byte {
	var err error
	c, d := getUapi(s.iface, s.logger, s.errs)

	serverIP := s.pool.GetStartingIP()
	clientIP, err := s.pool.Next(s.logger)
	if err != nil {
		s.logger.Errorf("pool.Next error: %v", err)
		s.errs <- err
	}

	peerKey, _ := wgtypes.GeneratePrivateKey()
	err = c.ConfigureDevice(s.iface, wgtypes.Config{
		ReplacePeers: false,
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:                   peerKey.PublicKey(),
				Remove:                      false,
				UpdateOnly:                  false,
				PersistentKeepaliveInterval: &s.duration,
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
					PersistentKeepaliveInterval: &s.duration,
					ReplaceAllowedIPs:           true,
					AllowedIPs: []net.IPNet{{
						IP:   serverIP,
						Mask: s.pool.CurrentIPMask(),
					}},
				},
			}},
		Address: []net.IPNet{{IP: clientIP,
			Mask: net.CIDRMask(32, 32)}},
	}
	subdomains[subdomain] = &wgQuickConfig
	configs, err := wgQuickConfig.MarshalText()
	s.logger.Verbosef("wgQuickConfig\n%s\n", configs)
	configString, _ := wgQuickConfig.MarshalText()
	return configString
}

func (s *WGServer) StopServer() {
	_ = s.uapiListen.Close()
	_ = s.createdTun.Close()
	s.createdDevice.Close()
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
