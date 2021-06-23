package util

import (
	"errors"
	"golang.zx2c4.com/wireguard/device"
	"net"
)

type IPPool struct {
	ipMask     net.IPMask
	startingIP net.IP
	currentIP  net.IP
}

func NewIPPool(cidr string) (*IPPool, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	} else {
		return &IPPool{
				ipMask:     dupMask(ipNet.Mask),
				startingIP: dupIP(ip),
				currentIP:  dupIP(ip),
			},
			nil
	}
}

func (i *IPPool) GetStartingIP() net.IP {
	return dupIP(i.startingIP)
}

func (i *IPPool) CurrentIP() net.IP {
	return dupIP(i.currentIP)
}

func (i *IPPool) CurrentIPMask() net.IPMask {
	return dupMask(i.ipMask)
}

func (i *IPPool) IPNet() *net.IPNet {
	return &net.IPNet{
		IP:   i.GetStartingIP(),
		Mask: i.CurrentIPMask(),
	}
}

func (i *IPPool) Next(logger *device.Logger) (net.IP, error) {
	clientIp := i.CurrentIP()
	logger.Verbosef("CURRENT IP %s", clientIp.String())
	for i := len(clientIp) - 1; i >= 0; i-- {
		logger.Verbosef("FOR i=%d clientIp[i]=%d", i, clientIp[i])
		clientIp[i]++
		if clientIp[i] < 255 && clientIp[i] > 0 {
			break
		} else {
			clientIp[i] = 1
		}
	}

	if !i.IPNet().Contains(clientIp) {
		logger.Verbosef("clientIp overflow i=%d clientIp=%v", i, clientIp)
		return nil, errors.New("clientIp overflow")
	} else {
		logger.Verbosef("new currentIP i=%d clientIp=%v", i, clientIp)
		i.currentIP = dupIP(clientIp)
		return clientIp, nil
	}
}

func dupIP(ip net.IP) net.IP {
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}
func dupMask(ip net.IPMask) net.IPMask {
	dup := make(net.IPMask, len(ip))
	copy(dup, ip)
	return dup
}
