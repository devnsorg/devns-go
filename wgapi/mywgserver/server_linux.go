// +build darwin

package mywgserver

func (s *WGServer) configureServerIP() {
	serverIP := s.pool.GetStartingIP()
	serverIPNet := net.IPNet{
		IP:   serverIP,
		Mask: s.pool.CurrentIPMask(),
	}
	println("ifconfig", strings.Split(fmt.Sprintf("%s inet %s %s alias", s.iface, serverIPNet.String(), serverIP.String()), " ")...)
	println("route", strings.Split(fmt.Sprintf("-q -n add -inet %s -interface %s", serverIPNet.String(), s.iface), " ")...)
}
