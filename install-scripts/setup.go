package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {
	// Detect the operating system
	osName := runtime.GOOS
	fmt.Println("Detected operating system:", osName)

	// Install package managers if not installed
	if osName == "darwin" {
		installHomebrew()
	} else if osName == "windows" {
		installChocolatey()
	}

	// Install Docker and Docker Compose based on the operating system
	if osName == "linux" {
		// Detect the Linux distribution
		distro := detectLinuxDistro()
		fmt.Println("Detected Linux distribution:", distro)

		// Install Docker and Docker Compose based on the distribution
		installDockerLinux(distro)
		installDockerComposeLinux(distro)
	} else if osName == "darwin" {
		// Install Docker Desktop for macOS
		installDockerMac()
		installDockerComposeMac()
	} else if osName == "windows" {
		// Install Docker Desktop for Windows
		installDockerWindows()
		installDockerComposeWindows()
	} else {
		fmt.Println("Unsupported operating system:", osName)
		return
	}

	// Build the Dockerfile
	buildDockerfile()

	// Run the application using Docker Compose
	runDockerCompose()

	fmt.Println("Application is running!")
}

func detectLinuxDistro() string {
	// Detect the Linux distribution based on the /etc/os-release file
	// You can add more distributions and their detection logic here
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		fmt.Println("Error reading /etc/os-release file:", err)
		return "unknown"
	}

	if strings.Contains(string(content), "Ubuntu") {
		return "ubuntu"
	} else if strings.Contains(string(content), "Arch") {
		return "arch"
	} else if strings.Contains(string(content), "Fedora") {
		return "fedora"
	} else if strings.Contains(string(content), "Debian") {
		return "debian"
	} else if strings.Contains(string(content), "Gentoo") {
		return "gentoo"
	} else if strings.Contains(string(content), "Slackware") {
		return "slackware"
	}

	return "unknown"
}

func installDockerLinux(distro string) {
	var installCmd *exec.Cmd
	var installBuildx *exec.Cmd

	switch distro {
	case "ubuntu", "debian":
		installCmd = exec.Command("sudo", "apt-get", "update")
		runCommand(installCmd)
		installCmd = exec.Command("sudo", "apt-get", "install", "-y", "docker.io")
	case "arch":
		installCmd = exec.Command("yay", "-S", "--noconfirm", "docker")
		installBuildx = exec.Command("sudo", "pacman", "-S", "--noconfirm", "docker-buildx")
	case "fedora":
		installCmd = exec.Command("sudo", "dnf", "install", "-y", "docker")
	case "gentoo":
		installCmd = exec.Command("sudo", "emerge", "--ask", "docker")
	case "slackware":
		installCmd = exec.Command("sudo", "slackpkg", "install", "docker")
	default:
		fmt.Println("Unsupported Linux distribution for Docker installation:", distro)
		return
	}

	runCommand(installCmd)
	runCommand(installBuildx)
	fmt.Println("Docker installed successfully.")
}

func installDockerComposeLinux(distro string) {
	var installCmd *exec.Cmd

	switch distro {
	case "ubuntu", "debian":
		installCmd = exec.Command("sudo", "apt-get", "install", "-y", "docker-compose")
	case "arch":
		installCmd = exec.Command("sudo", "pacman", "-Syu", "--noconfirm", "docker-compose")
	case "fedora":
		installCmd = exec.Command("sudo", "dnf", "install", "-y", "docker-compose")
	case "gentoo":
		installCmd = exec.Command("sudo", "emerge", "--ask", "docker-compose")
	case "slackware":
		installCmd = exec.Command("sudo", "slackpkg", "install", "docker-compose")
	default:
		fmt.Println("Unsupported Linux distribution for Docker Compose installation:", distro)
		return
	}

	runCommand(installCmd)
	fmt.Println("Docker Compose installed successfully.")
}

func installDockerMac() {
	// Install Docker Desktop for macOS using Homebrew
	installCmd := exec.Command("brew", "install", "--cask", "docker")
	runCommand(installCmd)
	fmt.Println("Docker Desktop installed successfully.")
}

func installDockerComposeMac() {
	// Docker Compose is included with Docker Desktop for macOS
	fmt.Println("Docker Compose is already installed with Docker Desktop.")
}

func installDockerWindows() {
	// Install Docker Desktop for Windows using Chocolatey
	installCmd := exec.Command("choco", "install", "docker-desktop", "-y")
	runCommand(installCmd)
	fmt.Println("Docker Desktop installed successfully.")
}

func installDockerComposeWindows() {
	// Docker Compose is included with Docker Desktop for Windows
	fmt.Println("Docker Compose is already installed with Docker Desktop.")
}

func buildDockerfile() {
	// Build the Dockerfile
	buildCmd := exec.Command("docker", "build", "-t", "myapp", ".")
	runCommand(buildCmd)
	fmt.Println("Dockerfile built successfully.")
}

func runDockerCompose() {
	// Run the application using Docker Compose
	runCmd := exec.Command("docker-compose", "up", "-d")
	runCommand(runCmd)
	fmt.Println("Application is running with Docker Compose.")
}

func runCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}
}

func installHomebrew() {
	// Check if Homebrew is installed
	_, err := exec.LookPath("brew")
	if err != nil {
		// Install Homebrew
		installCmd := exec.Command("/bin/bash", "-c", "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)")
		runCommand(installCmd)
		fmt.Println("Homebrew installed successfully.")
	} else {
		fmt.Println("Homebrew is already installed.")
	}
}

func installChocolatey() {
	// Check if Chocolatey is installed
	_, err := exec.LookPath("choco")
	if err != nil {
		// Install Chocolatey
		installCmd := exec.Command("powershell", "-Command", "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))")
		runCommand(installCmd)
		fmt.Println("Chocolatey installed successfully.")
	} else {
		fmt.Println("Chocolatey is already installed.")
	}
}
