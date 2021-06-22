// +build !darwin

package mywgclient

import (
	"fmt"
	"github.com/devnsorg/devns-go/wgclient/util"
	"net"
)

func (s *WGClient) configureServerIP() {
	allowedIPNet := s.wgQuickConfig.Peers[0].AllowedIPs[0]
	allowedIPNet.IP = allowedIPNet.IP.Mask(allowedIPNet.Mask)

	clientIPNet := net.IPNet{
		IP:   s.wgQuickConfig.Address[0].IP,
		Mask: allowedIPNet.Mask,
	}

	fmt.Printf("clientIPNet %s allowedIPNet %s", clientIPNet, allowedIPNet)
	util.ExecCommand(fmt.Sprintf("ip link set dev %s up", s.iface), s.logger)
	util.ExecCommand(fmt.Sprintf("ip addr add %s dev %s", clientIPNet.String(), s.iface), s.logger)
	util.ExecCommand(fmt.Sprintf("ip route add %s via %s dev %s", allowedIPNet.String(), clientIPNet.IP.String(), s.iface), s.logger)
}
