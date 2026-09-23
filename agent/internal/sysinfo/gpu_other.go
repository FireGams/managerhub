//go:build !linux && !windows && !darwin

package sysinfo

import "github.com/managerhub/managerhub/shared/protocol"

func detectGPUsPlatform() []protocol.GPUInfo { return nil }
