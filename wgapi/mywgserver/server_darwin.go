// +build darwin

package mywgserver

import (
	"fmt"
	"github.com/devnsorg/devns-go/util"
	"net"
)

func (s *WGServer) configureServerIP() {
	serverIP := s.ipPool.GetStartingIP()
	serverIPNet := net.IPNet{
		IP:   serverIP,
		Mask: s.ipPool.CurrentIPMask(),
	}
	_, _ = util.ExecCommand(fmt.Sprintf("ifconfig %s inet %s %s alias", s.iface, serverIPNet.String(), serverIP.String()), s.logger)
	_, _ = util.ExecCommand(fmt.Sprintf("route -q -n add -inet %s -interface %s", serverIPNet.String(), s.iface), s.logger)
}
