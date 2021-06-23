package util

import (
	"golang.zx2c4.com/wireguard/device"
	"os/exec"
	"strings"
)

func ExecCommand(command string, logger *device.Logger) ([]byte, error) {
	logger.Verbosef("EXEC COMMAND %s ", command)
	splitCmd := strings.Split(command, " ")
	ifconfig := exec.Command(splitCmd[0], splitCmd[1:]...)
	res, err := ifconfig.CombinedOutput()
	logger.Verbosef("EXEC RESPONSE %s ", string(res))
	return res, err
}
