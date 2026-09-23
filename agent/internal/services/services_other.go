//go:build !linux && !windows && !darwin

package services

import "github.com/managerhub/managerhub/shared/protocol"

type none struct{}

func newManager() Manager { return none{} }

func (none) List() ([]protocol.ServiceInfo, error) { return nil, nil }
func (none) Action(_, _ string) error              { return nil }
