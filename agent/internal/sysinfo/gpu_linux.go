//go:build linux

package sysinfo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/managerhub/managerhub/shared/protocol"
)

// detectGPUsPlatform scans /sys/class/drm for GPU devices.
func detectGPUsPlatform() []protocol.GPUInfo {
	var out []protocol.GPUInfo
	matches, _ := filepath.Glob("/sys/class/drm/card*/device/vendor")
	for _, m := range matches {
		b, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		vendor := strings.TrimSpace(string(b))
		g := protocol.GPUInfo{Vendor: vendorID(vendor)}
		if model, err := os.ReadFile(filepath.Join(filepath.Dir(m), "label")); err == nil {
			g.Model = strings.TrimSpace(string(model))
		}
		out = append(out, g)
	}
	return out
}

func vendorID(v string) string {
	switch strings.ToLower(v) {
	case "0x10de":
		return "nvidia"
	case "0x1002", "0x1022":
		return "amd"
	case "0x8086":
		return "intel"
	default:
		return v
	}
}
