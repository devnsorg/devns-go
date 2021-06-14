// +build linux

package mywgclient

import "fmt"

func (s *WGClient) configureServerIP() {
	clientIP := s.wgQuickConfig.Address[0].IP
	clientIPNet := s.wgQuickConfig.Address[0]
	fmt.Printf("clientIP %s clientIPNet %s", clientIP, clientIPNet)
}
