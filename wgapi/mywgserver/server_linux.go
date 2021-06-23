// +build !darwin

package mywgserver

import (
	"fmt"
	"github.com/devnsorg/devns-go/util"
	"net"
)

func (s *WGServer) configureServerIP() {
	serverIPNet := net.IPNet{
		IP:   s.ipPool.GetStartingIP(),
		Mask: s.ipPool.CurrentIPMask(),
	}

	allowedIPNet := net.IPNet{
		IP:   s.ipPool.GetStartingIP().Mask(s.ipPool.CurrentIPMask()),
		Mask: s.ipPool.CurrentIPMask(),
	}

	fmt.Printf("serverIPNet %s allowedIPNet %s", serverIPNet, allowedIPNet)
	util.ExecCommand(fmt.Sprintf("ip link set dev %s up", s.iface), s.logger)
	util.ExecCommand(fmt.Sprintf("ip addr add %s dev %s", serverIPNet.String(), s.iface), s.logger)
	util.ExecCommand(fmt.Sprintf("ip route add %s via %s dev %s", allowedIPNet.String(), serverIPNet.IP.String(), s.iface), s.logger)
}
