// +build darwin

package mywgclient

import (
	"fmt"
	"github.com/devnsorg/devns-go/util"
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
	util.ExecCommand(fmt.Sprintf("ifconfig %s inet %s %s alias", s.iface, clientIPNet.String(), clientIPNet.IP.String()), s.logger)
	util.ExecCommand(fmt.Sprintf("route -q -n add -inet %s -interface %s", allowedIPNet.String(), s.iface), s.logger)
}
