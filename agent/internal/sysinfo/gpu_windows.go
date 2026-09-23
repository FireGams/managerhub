//go:build windows

package sysinfo

import "github.com/managerhub/managerhub/shared/protocol"

// detectGPUsPlatform is a stub; full WMI query comes in a later milestone.
func detectGPUsPlatform() []protocol.GPUInfo { return nil }
