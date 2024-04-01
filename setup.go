package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

func main() {
	// Detect the operating system
	osName := runtime.GOOS
	fmt.Println("Detected operating system:", osName)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Install package managers if not installed
		if osName == "darwin" {
			installHomebrew(&wg)
		} else if osName == "windows" {
			installChocolatey(&wg)
		}
	}()

	// Create the './plex/config.py' file
	wg.Add(1)
	go func() {
		defer wg.Done()
		os.MkdirAll("./plex", os.ModePerm)
		file, err := os.Create("./plex/config.py")
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()
		file.WriteString(`HOST="192.168.1.1"
USER="root"
PASSWD="ThisIsAPassword"
DB="highseas"
IP="192.168.1.1"
PORT="32400"
`)
	}()

	// Create the environment files
	wg.Add(1)
	go func() {
		defer wg.Done()
		os.MkdirAll("./web/src/app/environments", os.ModePerm)
		environmentFiles := []string{"environment.prod.ts", "environment.ts", "environment.deployment.ts"}
		for _, fileName := range environmentFiles {
			file, err := os.Create("./web/src/app/environments/" + fileName)
			if err != nil {
				fmt.Println("Error creating file:", err)
				continue
			}
			defer file.Close()
			file.WriteString(`export const environment = {
    production: true,
    baseUrl: 'http://www.example.com:8080',
    envVar: {
      /**
       * Add environment variables you want to retriev from process
       * PORT:4200,
       * VAR_NAME: defaultValue
       */
      authorization: "THE BEARER TOKEN FOR TMDb API",
      port: "THE PORT YOUR GOLANG API IS RUNNING ON",
      ip: "THE IP YOUR GOLANG API IS RUNNING ON",
    },
  };
`)
		}
	}()

	// Create the './web/nginx.conf' file
	wg.Add(1)
	go func() {
		defer wg.Done()
		os.MkdirAll("./web", os.ModePerm)
		file, err := os.Create("./web/nginx.conf")
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()
		file.WriteString(`# the events block is required
events{}
http {
    # include the default mime.types to map file extensions to MIME types
    include /etc/nginx/mime.types;
    server {
        # set the root directory for the server (we need to copy our
        # application files here)
        root /usr/share/nginx/html;
        # set the default index file for the server (Angular generates the
        # index.html file for us and it will be in the above directory)
        index index.html;
        listen       6969;
        server_name http://goose.duocore.space http://arch.duocore.space;
        # specify the configuration for the '/' location
        location / {
            # try to serve the requested URI. if that fails then try to
            # serve the URI with a trailing slash. if that fails, then
            # serve the index.html file; this is needed in order to serve
            # Angular routes--e.g.,'localhost:8080/customer' will serve
            # the index.html file
            try_files $uri $uri/ /index.html;
        }
    }
}
`)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Install Docker and Docker Compose based on the operating system
		if osName == "linux" {
			// Detect the Linux distribution
			distro := detectLinuxDistro(&wg)
			fmt.Println("Detected Linux distribution:", distro)

			// Install Docker and Docker Compose based on the distribution
			installDockerLinux(distro, &wg)
			installDockerComposeLinux(distro, &wg)
		} else if osName == "darwin" {
			// Install Docker Desktop for macOS
			installDockerMac(&wg)
			installDockerComposeMac(&wg)
		} else if osName == "windows" {
			// Install Docker Desktop for Windows
			installDockerWindows(&wg)
			installDockerComposeWindows(&wg)
		} else {
			fmt.Println("Unsupported operating system:", osName)
			return
		}
	}()

	go func() {
		defer wg.Done()

		// Build the Dockerfile
		buildDockerfileWeb(&wg)
		buildDockerfileBackendGo(&wg)
		buildDockerfileBackendPython(&wg)

		// Run the application using Docker Compose
		runDockerCompose(&wg)
	}()

	wg.Wait()

	fmt.Println("Application is running!")
}

func detectLinuxDistro(wg *sync.WaitGroup) string {
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

func installDockerLinux(distro string, wg *sync.WaitGroup) {
	var installCmd *exec.Cmd

	switch distro {
	case "ubuntu", "debian":
		installCmd = exec.Command("sudo", "apt-get", "update")
		runCommand(installCmd, wg)
		installCmd = exec.Command("sudo", "apt-get", "install", "-y", "docker.io")
	case "arch":
		installCmd = exec.Command("sudo", "pacman", "-Syu", "--noconfirm", "docker")
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

	runCommand(installCmd, wg)
	fmt.Println("Docker installed successfully.")
}

func installDockerComposeLinux(distro string, wg *sync.WaitGroup) {
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

	runCommand(installCmd, wg)
	fmt.Println("Docker Compose installed successfully.")
}

func installDockerMac(wg *sync.WaitGroup) {
	// Install Docker Desktop for macOS using Homebrew
	installCmd := exec.Command("brew", "install", "--cask", "docker")
	runCommand(installCmd, wg)
	fmt.Println("Docker Desktop installed successfully.")
}

func installDockerComposeMac(wg *sync.WaitGroup) {
	// Docker Compose is included with Docker Desktop for macOS
	fmt.Println("Docker Compose is already installed with Docker Desktop.")
}

func installDockerWindows(wg *sync.WaitGroup) {
	// Install Docker Desktop for Windows using Chocolatey
	_, err := exec.LookPath("docker-desktop")
	if err != nil {
		// Install Chocolatey
		installCmd := exec.Command("choco", "install", "-y", "docker-desktop")
		runCommand(installCmd, wg)
		fmt.Println("Docker Desktop installed successfully.")
	} else {
		fmt.Println("Docker Desktop is already installed.")
	}

}

func installDockerComposeWindows(wg *sync.WaitGroup) {
	// Docker Compose is included with Docker Desktop for Windows
	fmt.Println("Docker Compose is already installed with Docker Desktop.")
}

func buildDockerfileWeb(wg *sync.WaitGroup) {
	// Build the Dockerfile
	changeDir := exec.Command("cd", "./web")
	runCommand(changeDir, wg)
	buildCmd := exec.Command("docker", "build", "-t", "high-seas-frontend", ".")
	runCommand(buildCmd, wg)
	fmt.Println("Dockerfile for High Seas frontend built successfully.")
}

func buildDockerfileBackendGo(wg *sync.WaitGroup) {
	// Build the Dockerfile
	buildCmd := exec.Command("docker", "build", "-t", "high-seas-golang", ".")
	runCommand(buildCmd, wg)
	fmt.Println("Dockerfile for High Seas backend in Golang built successfully.")
}

func buildDockerfileBackendPython(wg *sync.WaitGroup) {
	// Build the Dockerfile
	buildCmd := exec.Command("docker", "build", "-t", "high-seas-python", ".")
	runCommand(buildCmd, wg)
	fmt.Println("Dockerfile for High Seas backend in Python built successfully.")
}

func runDockerCompose(wg *sync.WaitGroup) {
	// Run the application using Docker Compose
	runCmd := exec.Command("docker-compose", "up", "-d")
	runCommand(runCmd, wg)
	fmt.Println("Application is running with Docker Compose.")
}

func runCommand(cmd *exec.Cmd, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	} else {
		fmt.Printf("Command executed successfully")
	}
}

func installHomebrew(wg *sync.WaitGroup) {
	// Check if Homebrew is installed
	_, err := exec.LookPath("brew")
	if err != nil {
		// Install Homebrew
		installCmd := exec.Command("/bin/bash", "-c", "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)")
		runCommand(installCmd, wg)
		fmt.Println("Homebrew installed successfully.")
	} else {
		fmt.Println("Homebrew is already installed.")
	}
}

func installChocolatey(wg *sync.WaitGroup) {
	// Check if Chocolatey is installed
	_, err := exec.LookPath("choco")
	if err != nil {
		// Install Chocolatey
		installCmd := exec.Command("powershell", "-Command", "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))")
		runCommand(installCmd, wg)
		fmt.Println("Chocolatey installed successfully.")
	} else {
		fmt.Println("Chocolatey is already installed.")
	}
}
