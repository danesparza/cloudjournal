package system

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// GetHostname returns the contents of /etc/hostname
func GetHostname() string {

	retval, err := os.ReadFile("/etc/hostname")
	if err != nil {
		log.WithError(err).Error("problem getting hostname")
		return ""
	}

	return strings.TrimSpace(string(retval))
}

// GetMachineID returns the contents of /etc/machine-id
func GetMachineID() string {
	retval, err := os.ReadFile("/etc/machine-id")
	if err != nil {
		log.WithError(err).Error("problem getting machine id")
		return ""
	}

	return strings.TrimSpace(string(retval))
}
