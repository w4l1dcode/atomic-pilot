package atomic

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// CheckAtomicTestExists verifies if an Atomic Red Team test exists for a given technique.
func CheckAtomicTestExists(technique string, installPath string) (bool, string) {
	command := fmt.Sprintf("Import-Module \"%s\\invoke-atomicredteam\\Invoke-AtomicRedTeam.psd1\" -Force; Invoke-AtomicTest %s -ShowDetailsBrief", installPath, technique)
	cmd := exec.Command("pwsh", "-Command", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error checking Atomic Red Team test for technique %s:\n%s", technique, string(output))
		return false, ""
	}

	outputStr := string(output)

	// Check for the specific error message indicating the test does not exist
	if strings.Contains(outputStr, "does not exist") {
		return false, ""
	}

	// Check for messages indicating no applicable tests for macOS or Windows
	if strings.Contains(outputStr, fmt.Sprintf("Found 0 atomic tests applicable to macos platform for Technique %s", technique)) {
		return false, "macOS"
	}

	if strings.Contains(outputStr, fmt.Sprintf("Found 0 atomic tests applicable to windows platform for Technique %s", technique)) {
		return false, "Windows"
	}

	return true, ""
}

func RunAtomicTest(technique string, installPath string) {
	// PowerShell command to execute the Atomic Red Team test
	command := fmt.Sprintf("Import-Module \"%s\\invoke-atomicredteam\\Invoke-AtomicRedTeam.psd1\" -Force; Invoke-AtomicTest %s", installPath, technique)
	cmd := exec.Command("pwsh", "-Command", command)

	// Execute the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Error running Atomic Red Team test for technique %s:\n%s", technique, string(output))
		return
	}
	fmt.Printf("Output of Atomic Red Team test for technique %s:\n%s\n", technique, string(output))
}

func FilterObjectsByType(objects []interface{}, objectType string) []interface{} {
	var filtered []interface{}
	for _, obj := range objects {
		objMap := obj.(map[string]interface{})
		if objMap["type"] == objectType {
			filtered = append(filtered, obj)
		}
	}
	return filtered
}

func GetObjectByID(objects []interface{}, id string) map[string]interface{} {
	for _, obj := range objects {
		objMap := obj.(map[string]interface{})
		if objMap["id"] == id {
			return objMap
		}
	}
	return nil
}
