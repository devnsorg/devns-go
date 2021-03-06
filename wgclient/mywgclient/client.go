package mywgclient

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

type WGClient struct {
	iface         string
	wgQuickConfig *util.WgQuickConfig
	logger        *device.Logger
	errs          chan error
	createdTun    tun.Device
	createdDevice *device.Device
	uapiListen    net.Listener
	duration      time.Duration
	wgPool        *util.WGPool
}

func NewWGClient(wgQuickConfigString string, logger *device.Logger, errs chan error) *WGClient {
	var err error
	wgQuickConfig := &util.WgQuickConfig{}
	err = wgQuickConfig.UnmarshalText([]byte(wgQuickConfigString))

	if err != nil {
		logger.Errorf("ERROR %#v", err)
		errs <- err
	}

	return &WGClient{wgQuickConfig: wgQuickConfig, logger: logger, errs: errs, duration: *wgQuickConfig.Peers[0].PersistentKeepaliveInterval}
}

func (s *WGClient) StartServer() chan struct{} {
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
		err := errors.New("error due to inactivity from peer")
		s.logger.Errorf("CleanUpStalePeers %#v", err)
		s.errs <- err
	})
	s.wgPool.AddPoolPeerByPubKey(s.wgQuickConfig.Peers[0].PublicKey)

	return createdDevice.Wait()
}

func (s *WGClient) createDevice() (string, tun.Device, *device.Device) {
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

func (s *WGClient) configureDevice() {
	var err error
	c, _ := util.GetUapi(s.iface, s.logger, s.errs)

	pk := s.wgQuickConfig.PrivateKey

	err = c.ConfigureDevice(s.iface, wgtypes.Config{
		PrivateKey:   pk,
		FirewallMark: nil,
		ReplacePeers: true,
		Peers:        s.wgQuickConfig.Peers,
	})

	if err != nil {
		s.logger.Errorf("ERROR %#v\n", err)
		s.errs <- err
	}
}

func (s *WGClient) StopClient() {
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
