package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Install downloads and configures a GitHub Actions runner.
func Install(repoURL, token, name, labels, workDir string) (string, error) {
	if workDir == "" {
		workDir = "/opt/actions-runner"
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	arch := "x64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}
	osName := "linux"
	switch runtime.GOOS {
	case "darwin":
		osName = "osx"
	case "windows":
		osName = "win"
	}
	version := "2.337.0"
	pkg := fmt.Sprintf("actions-runner-%s-%s-%s.tar.gz", osName, arch, version)
	url := fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/%s", version, pkg)

	var out strings.Builder
	out.WriteString(fmt.Sprintf("Downloading %s...\n", pkg))

	// Download
	dlPath := filepath.Join(workDir, pkg)
	if err := downloadFile(url, dlPath); err != nil {
		return out.String(), fmt.Errorf("download: %w", err)
	}

	// Extract
	out.WriteString("Extracting...\n")
	if runtime.GOOS == "windows" {
		if err := exec.Command("tar", "-xzf", dlPath, "-C", workDir).Run(); err != nil {
			return out.String(), fmt.Errorf("extract: %w", err)
		}
	} else {
		if err := exec.Command("tar", "-xzf", dlPath, "-C", workDir).Run(); err != nil {
			return out.String(), fmt.Errorf("extract: %w", err)
		}
	}
	_ = os.Remove(dlPath)

	// Configure
	out.WriteString(fmt.Sprintf("Configuring runner %s for %s...\n", name, repoURL))
	cfgArgs := []string{"--url", repoURL, "--token", token, "--name", name, "--unattended", "--replace"}
	if labels != "" {
		cfgArgs = append(cfgArgs, "--labels", labels)
	}
	cfgCmd := exec.Command(filepath.Join(workDir, "config.sh"), cfgArgs...)
	cfgCmd.Dir = workDir
	cfgOut, err := cfgCmd.CombinedOutput()
	out.Write(cfgOut)
	if err != nil {
		return out.String(), fmt.Errorf("configure: %w", err)
	}

	// Install as service
	out.WriteString("Installing service...\n")
	svcCmd := exec.Command(filepath.Join(workDir, "svc.sh"), "install")
	svcCmd.Dir = workDir
	svcOut, err := svcCmd.CombinedOutput()
	out.Write(svcOut)
	if err != nil {
		out.WriteString("Warning: service install failed (can run manually with run.sh)\n")
	} else {
		// Start service
		startCmd := exec.Command(filepath.Join(workDir, "svc.sh"), "start")
		startCmd.Dir = workDir
		startOut, _ := startCmd.CombinedOutput()
		out.Write(startOut)
		out.WriteString("Runner installed and started!\n")
	}

	return out.String(), nil
}

func downloadFile(url, dest string) error {
	return exec.Command("curl", "-fsSL", "-o", dest, url).Run()
}
