//go:build darwin

package sysinfo

import "github.com/managerhub/managerhub/shared/protocol"

// detectGPUsPlatform is a stub; system_profiler query comes later.
func detectGPUsPlatform() []protocol.GPUInfo { return nil }
