package utils
import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)
type Dependency struct {
	Name        string
	Command     string
	Required    bool
	InstallHint string
	CheckFunc   func() error
}
type DependencyManager struct {
	dependencies []Dependency
}
func NewDependencyManager() *DependencyManager {
	dm := &DependencyManager{}
	dm.initializeDependencies()
	return dm
}
func (dm *DependencyManager) initializeDependencies() {
	dm.dependencies = []Dependency{
		{
			Name:        "OpenVPN",
			Command:     "openvpn",
			Required:    true,
			InstallHint: "Install OpenVPN: sudo apt-get install openvpn (Ubuntu/Debian) or sudo yum install openvpn (RHEL/CentOS)",
			CheckFunc:   dm.checkOpenVPN,
		},
		{
			Name:        "EasyRSA",
			Command:     "easyrsa",
			Required:    true,
			InstallHint: "Install EasyRSA: sudo apt-get install easy-rsa (Ubuntu/Debian) or download from GitHub",
			CheckFunc:   dm.checkEasyRSA,
		},
		{
			Name:        "SystemD",
			Command:     "systemctl",
			Required:    true,
			InstallHint: "SystemD should be available on modern Linux systems",
			CheckFunc:   dm.checkSystemD,
		},
		{
			Name:        "Python3",
			Command:     "python3",
			Required:    false,
			InstallHint: "Install Python3: sudo apt-get install python3 (Ubuntu/Debian)",
			CheckFunc:   dm.checkPython3,
		},
	}
}
func (dm *DependencyManager) ValidateAll() error {
	var errors []string
	for _, dep := range dm.dependencies {
		if err := dep.CheckFunc(); err != nil {
			if dep.Required {
				errors = append(errors, fmt.Sprintf("❌ %s: %s\n   💡 %s", dep.Name, err.Error(), dep.InstallHint))
			} else {
				fmt.Printf("⚠️  %s: %s (optional)\n", dep.Name, err.Error())
			}
		} else {
			fmt.Printf("✅ %s: Available\n", dep.Name)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("missing required dependencies:\n%s", strings.Join(errors, "\n"))
	}
	return nil
}
func (dm *DependencyManager) CheckDependency(name string) error {
	for _, dep := range dm.dependencies {
		if dep.Name == name {
			return dep.CheckFunc()
		}
	}
	return fmt.Errorf("unknown dependency: %s", name)
}
func (dm *DependencyManager) checkOpenVPN() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("OpenVPN is only supported on Linux")
	}
	if _, err := exec.LookPath("openvpn"); err != nil {
		return fmt.Errorf("openvpn command not found")
	}
	cmd := exec.Command("openvpn", "--version")
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get OpenVPN version: %v", err)
	}
	return nil
}
func (dm *DependencyManager) checkEasyRSA() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("EasyRSA is only supported on Linux")
	}
	if _, err := exec.LookPath("easyrsa"); err != nil {
		return fmt.Errorf("easyrsa command not found")
	}
	easyRSAPath := "/etc/openvpn/easy-rsa"
	if _, err := os.Stat(easyRSAPath); os.IsNotExist(err) {
		return fmt.Errorf("EasyRSA directory not found at %s", easyRSAPath)
	}
	return nil
}
func (dm *DependencyManager) checkSystemD() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("SystemD is only available on Linux")
	}
	if _, err := exec.LookPath("systemctl"); err != nil {
		return fmt.Errorf("systemctl command not found")
	}
	cmd := exec.Command("systemctl", "is-system-running")
	output, err := cmd.Output()
	if err != nil {
		// systemctl is-system-running can return non-zero for degraded/starting states
		// Check if output indicates system is functional
		outputStr := strings.TrimSpace(string(output))
		if outputStr == "degraded" || outputStr == "starting" || outputStr == "initializing" {
			// These states are acceptable - systemd is running
			return nil
		}
		return fmt.Errorf("systemd not running or accessible: %v", err)
	}
	// Check if output indicates system is functional
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "running" || outputStr == "degraded" || outputStr == "starting" {
		return nil
	}
	return fmt.Errorf("systemd is not in a functional state: %s", outputStr)
}
func (dm *DependencyManager) checkPython3() error {
	if _, err := exec.LookPath("python3"); err != nil {
		return fmt.Errorf("python3 command not found")
	}
	cmd := exec.Command("python3", "--version")
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get Python version: %v", err)
	}
	return nil
}
func (dm *DependencyManager) GetMissingDependencies() []string {
	var missing []string
	for _, dep := range dm.dependencies {
		if dep.Required {
			if err := dep.CheckFunc(); err != nil {
				missing = append(missing, dep.Name)
			}
		}
	}
	return missing
}