package util

import (
	"errors"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"time"
)

type WGPoolPeer struct {
	subdomain        string
	address          net.IP
	publicKey        wgtypes.Key
	lastReceiveBytes int64
}

func (W *WGPoolPeer) PublicKey() wgtypes.Key {
	return W.publicKey
}

type WGPool struct {
	subdomains map[string]*WGPoolPeer
	iface      string
	logger     *device.Logger
	errs       chan error
}

type onStale func(poolPeer *WGPoolPeer)

func NewWGPool(iface string, logger *device.Logger, errs chan error) *WGPool {
	return &WGPool{subdomains: make(map[string]*WGPoolPeer), iface: iface, logger: logger, errs: errs}
}

func (w *WGPool) AddPoolPeer(subdomain string, publicKey wgtypes.Key, ipAddress net.IP) {
	w.subdomains[subdomain] = &WGPoolPeer{
		subdomain:        subdomain,
		address:          ipAddress,
		publicKey:        publicKey,
		lastReceiveBytes: -10, // To skip the first inactivity check
	}
}
func (w *WGPool) AddPoolPeerByPubKey(publicKey wgtypes.Key) {
	w.subdomains["default"] = &WGPoolPeer{
		publicKey:        publicKey,
		lastReceiveBytes: -10, // To skip the first inactivity check
	}
}

func (w *WGPool) GetPeerAddressBySubdomain(subdomain string) net.IP {
	return w.subdomains[subdomain].address
}

func (w *WGPool) GetPoolPeerByPubKey(pubkey wgtypes.Key) *WGPoolPeer {
	for iSubdomain, config := range w.subdomains {
		if config.publicKey == pubkey {
			w.logger.Verbosef("GetPoolPeerByPubKey found %s", iSubdomain)
			return config
		}
	}
	w.errs <- errors.New("GetPeerAddressBySubdomain not found")
	return nil
}

func (w *WGPool) CleanUpStalePeers(keepAliveDuration time.Duration, onStaleF onStale) {
	for range time.Tick(2 * keepAliveDuration) {
		_, d := GetUapi(w.iface, w.logger, w.errs)
		for _, peer := range d.Peers {
			poolPeer := w.GetPoolPeerByPubKey(peer.PublicKey)

			//COMPARE BYTES
			if peer.ReceiveBytes == poolPeer.lastReceiveBytes {
				delete(w.subdomains, poolPeer.subdomain)
				// CALL onStale
				onStaleF(poolPeer)
				w.logger.Verbosef("Peer removed from POOL %#v", poolPeer)
			}

			//UPDATE BYTES
			poolPeer.lastReceiveBytes = peer.ReceiveBytes
		}
	}
}

//func cleanUpStalePeers(iface string, duration time.Duration, logger *device.Logger, errs chan error) {
//	for range time.Tick(duration * 2) {
//		c, d := GetUapi(iface, logger, errs)
//		for _, peer := range d.Peers {
//			var subdomain = ""
//			for iSubdomain, config := range subdomains {
//				if config.PrivateKey.PublicKey() == peer.PublicKey {
//					logger.Verbosef("REMOVE PEER Map %s", iSubdomain)
//					subdomain = iSubdomain
//				}
//			}
//			if len(subdomain) == 0 {
//				errs <- errors.New("Map not synced")
//			}
//
//			if peer.LastHandshakeTime.IsZero() && !subdomainsZeroHandshake[subdomain] {
//				subdomainsZeroHandshake[subdomain] = true
//			} else if peer.LastHandshakeTime.Add(2 * duration).Before(time.Now()) {
//				// If 2xDURATION passed, delete peer
//				err := c.ConfigureDevice(iface, wgtypes.Config{
//					ReplacePeers: false,
//					Peers: []wgtypes.PeerConfig{{
//						PublicKey:  peer.PublicKey,
//						Remove:     true,
//						UpdateOnly: true,
//					}},
//				})
//				if err != nil {
//					logger.Errorf("REMOVE PEER ConfigureDevice %#v", err)
//					errs <- err
//				}
//				delete(subdomains, subdomain)
//				delete(subdomainsZeroHandshake, subdomain)
//			}
//		}
//	}
//}
