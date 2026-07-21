package deps

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func checkDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func installPackage(pkg string) error {
	fmt.Printf("📦 Installing %s...\n", pkg)

	var cmd *exec.Cmd
	if checkDependency("apt-get") {
		cmd = exec.Command("sudo", "apt-get", "install", "-y", pkg)
	} else if checkDependency("dnf") {
		cmd = exec.Command("sudo", "dnf", "install", "-y", pkg)
	} else if checkDependency("pacman") {
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", pkg)
	} else if checkDependency("brew") {
		cmd = exec.Command("brew", "install", pkg)
	} else {
		return fmt.Errorf("no supported package manager found (apt, dnf, pacman, brew)")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func promptYesNo(message string) bool {
	fmt.Print(message)

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

func Ensure() {
	ensureYTDLP()

	deps := map[string]string{
		"mpv": "mpv",
		"fzf": "fzf",
	}

	missing := []string{}
	for cmd := range deps {
		if !checkDependency(cmd) {
			missing = append(missing, cmd)
		}
	}

	if len(missing) == 0 {
		return
	}

	fmt.Println("⚠️  Missing required dependencies:")
	for _, dep := range missing {
		fmt.Printf("   - %s\n", dep)
	}
	fmt.Println()
	if !promptYesNo("📥 Would you like to install them now? (y/N): ") {
		fmt.Println("❌ Cannot proceed without required dependencies")
		os.Exit(1)
	}

	fmt.Println()
	for _, dep := range missing {
		if err := installPackage(deps[dep]); err != nil {
			log.Fatalf("❌ Failed to install %s: %v\n", dep, err)
		}
		fmt.Printf("✅ %s installed successfully\n", dep)
	}
	fmt.Println()
}
