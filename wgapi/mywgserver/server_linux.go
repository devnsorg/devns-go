// +build !darwin

package mywgserver

import (
	"fmt"
	"github.com/ipTLS/ipTLS/wgapi/util"
	"net"
)

func (s *WGServer) configureServerIP() {
	serverIPNet := net.IPNet{
		IP:   s.pool.GetStartingIP(),
		Mask: s.pool.CurrentIPMask(),
	}

	allowedIPNet := net.IPNet{
		IP:   s.pool.GetStartingIP().Mask(s.pool.CurrentIPMask()),
		Mask: s.pool.CurrentIPMask(),
	}

	fmt.Printf("serverIPNet %s allowedIPNet %s", serverIPNet, allowedIPNet)
	util.ExecCommand(fmt.Sprintf("ip link set dev %s up", s.iface), s.logger)
	util.ExecCommand(fmt.Sprintf("ip addr add %s dev %s", serverIPNet.String(), s.iface), s.logger)
	util.ExecCommand(fmt.Sprintf("ip route add %s via %s dev %s", allowedIPNet.String(), serverIPNet.IP.String(), s.iface), s.logger)
}