// +build darwin

package mywgclient

import (
	"fmt"
	"os/exec"
	"strings"
)

func (s *WGClient) configureServerIP() {
	clientIP := s.wgQuickConfig.Address[0].IP
	clientIPNet := s.wgQuickConfig.Address[0]
	fmt.Printf("clientIP %s clientIPNet %s", clientIP, clientIPNet)

	ifconfig := exec.Command("ifconfig", strings.Split(fmt.Sprintf("%s inet %s %s alias", s.iface, clientIPNet.String(), clientIP.String()), " ")...)
	stdoutStderr, err := ifconfig.CombinedOutput()
	if err != nil {
		s.logger.Errorf("ERROR %#v", err)
		s.errs <- err
	}
	s.logger.Verbosef("%s\n", stdoutStderr)

	route := exec.Command("route", strings.Split(fmt.Sprintf("-q -n add -inet %s -interface %s", clientIPNet.String(), s.iface), " ")...)
	stdoutStderr, err = route.CombinedOutput()
	if err != nil {
		s.logger.Errorf("ERROR %#v", err)
		s.errs <- err
	}
	s.logger.Verbosef("%s\n", stdoutStderr)
}
