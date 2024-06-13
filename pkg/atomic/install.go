package atomic

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// GetInstallPath returns the installation path based on the OS
func GetInstallPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join("C:\\", "AtomicRedTeam"), nil
	default:
		return filepath.Join(homeDir, "AtomicRedTeam"), nil
	}
}

func InstallAndVerifyAtomicRedTeam(installPath string) error {
	fmt.Println("Installing Atomic Red Team tests...")

	script := fmt.Sprintf(`IEX (IWR 'https://raw.githubusercontent.com/redcanaryco/invoke-atomicredteam/master/install-atomicredteam.ps1' -UseBasicParsing); Install-AtomicRedTeam -getAtomics -Force; Import-Module "%s\invoke-atomicredteam\Invoke-AtomicRedTeam.psd1" -Force; $PSDefaultParameterValues = @{"Invoke-AtomicTest:PathToAtomicsFolder"="%s\atomics"}`, installPath, installPath)
	cmd := exec.Command("pwsh", "-Command", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error installing or importing Atomic Red Team: %v\n%s", string(output))
	}
	return nil
}
