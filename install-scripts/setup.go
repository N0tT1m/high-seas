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

	// Create the './plex/config.py' file
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

	// Create the environment files
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

	// Create the './web/nginx.conf' file

	os.MkdirAll("./web", os.ModePerm)
	file, err = os.Create("./web/nginx.conf")
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

	// Start Docker
	startDocker()

	// Pass the password to the runCommand function

	// Build the Dockerfile
	buildDockerfileWeb()
	buildDockerfileBackendGo()
	buildDockerfileBackendPython()

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
		installCmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", "docker-compose")
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
	_, err := exec.LookPath("docker-desktop")
	if err != nil {
		// Install Chocolatey
		installCmd := exec.Command("choco", "install", "-y", "docker-desktop")
		runCommand(installCmd)
		fmt.Println("Docker Desktop installed successfully.")
	} else {
		fmt.Println("Docker Desktop is already installed.")
	}

}

func installDockerComposeWindows() {
	// Docker Compose is included with Docker Desktop for Windows
	fmt.Println("Docker Compose is already installed with Docker Desktop.")
}

func buildDockerfileWeb() {
	// Build the Dockerfile
	osName := runtime.GOOS
	var buildCmd *exec.Cmd

	if osName == "linux" {
		distro := detectLinuxDistro()
		switch distro {
		case "ubuntu", "debian":
			buildCmd = exec.Command("sudo", "docker", "build", "-t", "high-seas-frontend", "./web")
		case "arch":
			buildCmd = exec.Command("sudo", "docker", "buildx", "build", "-t", "high-seas-frontend", "./web")
		case "fedora", "gentoo", "slackware":
			buildCmd = exec.Command("sudo", "docker", "build", "-t", "high-seas-frontend", "./web")
		default:
			fmt.Println("Unsupported Linux distribution for building Dockerfile:", distro)
			return
		}
	} else if osName == "darwin" {
		buildCmd = exec.Command("docker", "build", "-t", "high-seas-frontend", "./web")
	} else if osName == "windows" {
		buildCmd = exec.Command("docker", "build", "-t", "high-seas-frontend", "./web")
	} else {
		fmt.Println("Unsupported operating system for building Dockerfile:", osName)
		return
	}

	runCommand(buildCmd)
	fmt.Println("Dockerfile for High Seas frontend built successfully.")
}

func buildDockerfileBackendGo() {
	osName := runtime.GOOS
	var buildCmd *exec.Cmd

	if osName == "linux" {
		distro := detectLinuxDistro()
		switch distro {
		case "ubuntu", "debian":
			buildCmd = exec.Command("sudo", "docker", "build", "-t", "high-seas-golang", ".")
		case "arch":
			buildCmd = exec.Command("sudo", "docker", "buildx", "build", "-t", "high-seas-golang", ".")
		case "fedora", "gentoo", "slackware":
			buildCmd = exec.Command("sudo", "docker", "build", "-t", "high-seas-golang", ".")
		default:
			fmt.Println("Unsupported Linux distribution for building Dockerfile:", distro)
			return
		}
	} else if osName == "darwin" {
		buildCmd = exec.Command("docker", "build", "-t", "high-seas-golang", ".")
	} else if osName == "windows" {
		buildCmd = exec.Command("docker", "build", "-t", "high-seas-golang", ".")
	} else {
		fmt.Println("Unsupported operating system for building Dockerfile:", osName)
		return
	}

	runCommand(buildCmd)
	fmt.Println("Dockerfile for High Seas backend in Golang built successfully.")
}

func buildDockerfileBackendPython() {
	osName := runtime.GOOS
	var buildCmd *exec.Cmd

	if osName == "linux" {
		distro := detectLinuxDistro()
		switch distro {
		case "ubuntu", "debian":
			buildCmd = exec.Command("sudo", "docker", "build", "-t", "high-seas-python", ".")
		case "arch":
			buildCmd = exec.Command("sudo", "docker", "buildx", "build", "-t", "high-seas-python", ".")
		case "fedora", "gentoo", "slackware":
			buildCmd = exec.Command("sudo", "docker", "build", "-t", "high-seas-python", ".")
		default:
			fmt.Println("Unsupported Linux distribution for building Dockerfile:", distro)
			return
		}
	} else if osName == "darwin" {
		buildCmd = exec.Command("docker", "build", "-t", "high-seas-python", ".")
	} else if osName == "windows" {
		buildCmd = exec.Command("docker", "build", "-t", "high-seas-python", ".")
	} else {
		fmt.Println("Unsupported operating system for building Dockerfile:", osName)
		return
	}

	runCommand(buildCmd)
	fmt.Println("Dockerfile for High Seas backend in Python built successfully.")
}

func runDockerCompose() {
	osName := runtime.GOOS
	var runCmd *exec.Cmd

	if osName == "linux" {
		distro := detectLinuxDistro()
		switch distro {
		case "ubuntu", "debian":
			runCmd = exec.Command("sudo", "docker-compose", "up", "-d")
		case "arch":
			runCmd = exec.Command("sudo", "docker-compose", "up", "-d")
		case "fedora", "gentoo", "slackware":
			runCmd = exec.Command("sudo", "docker-compose", "up", "-d")
		default:
			fmt.Println("Unsupported Linux distribution for running Docker Compose:", distro)
			return
		}
	} else if osName == "darwin" {
		runCmd = exec.Command("docker-compose", "up", "-d")
	} else if osName == "windows" {
		runCmd = exec.Command("docker-compose", "up", "-d")
	} else {
		fmt.Println("Unsupported operating system for running Docker Compose:", osName)
		return
	}

	runCommand(runCmd)
	fmt.Println("Application is running with Docker Compose.")
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

// ... (rest of the functions remain the same)

func runCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a pipe to pass the password to sudo via stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error creating stdin pipe:", err)
		os.Exit(1)
	}

	// Start the command
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting command:", err)
		os.Exit(1)
	}

	stdin.Close()

	// Wait for the command to finish
	err = cmd.Wait()
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	} else {
		fmt.Printf("Command executed successfully\n")
	}
}

func startDocker() {
	osName := runtime.GOOS

	if osName == "linux" {
		if isSystemdUsed() {
			startDockerSystemd()
		} else {
			startDockerRcd()
		}
	} else if osName == "darwin" {
		startDockerMac()
	} else if osName == "windows" {
		startDockerWindows()
	} else {
		fmt.Println("Unsupported operating system for starting Docker:", osName)
		return
	}
}

func isSystemdUsed() bool {
	// Check if systemd is used as the init system
	_, err := os.Stat("/run/systemd/system")
	return err == nil
}

func startDockerSystemd() {
	startCmd := exec.Command("sudo", "systemctl", "start", "docker")
	runCommand(startCmd)
	fmt.Println("Docker started successfully using systemd.")
}

func startDockerRcd() {
	startCmd := exec.Command("sudo", "service", "docker", "start")
	runCommand(startCmd)
	fmt.Println("Docker started successfully using rc.d.")
}

func startDockerMac() {
	startCmd := exec.Command("open", "/Applications/Docker.app")
	runCommand(startCmd)
	fmt.Println("Docker started successfully on macOS.")
}

func startDockerWindows() {
	startCmd := exec.Command("C:\\Program Files\\Docker\\Docker\\Docker Desktop.exe")
	runCommand(startCmd)
	fmt.Println("Docker started successfully on Windows.")
}
