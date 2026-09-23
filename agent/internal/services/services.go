// Package services lists and controls host services (systemd, Windows, launchd).
package services

import "github.com/managerhub/managerhub/shared/protocol"

// Manager abstracts the host service manager.
type Manager interface {
	List() ([]protocol.ServiceInfo, error)
	Action(name, action string) error
}

// New returns the platform service manager.
func New() Manager { return newManager() }
