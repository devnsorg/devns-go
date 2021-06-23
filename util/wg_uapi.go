package util

import (
	"errors"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func GetUapi(iface string, logger *device.Logger, errs chan error) (*wgctrl.Client, *wgtypes.Device) {
	uapiClient, err := wgctrl.New()
	if err != nil {
		logger.Errorf("wgctrl error: %v", err)
		errs <- err
	}
	devices, err := uapiClient.Devices()
	if err != nil {
		logger.Errorf("wgctrl get Devices error: %v", err)
		errs <- err
	}

	var uapiDevice *wgtypes.Device
	for _, iDevice := range devices {
		if iDevice.Name == iface {
			uapiDevice = iDevice
		}
	}

	if uapiDevice == nil {
		err = errors.New("device not found")
		errs <- err
		return nil, nil
	}
	return uapiClient, uapiDevice
}
