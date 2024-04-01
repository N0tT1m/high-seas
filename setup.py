import os
import subprocess
import platform
import shutil

def main():
    # Detect the operating system
    os_name = platform.system()
    print("Detected operating system:", os_name)

    # Install package managers if not installed
    if os_name == "Darwin":
        install_homebrew()
    elif os_name == "Windows":
        install_chocolatey()

    # Install Docker and Docker Compose based on the operating system
    if os_name == "Linux":
        # Detect the Linux distribution
        distro = detect_linux_distro()
        print("Detected Linux distribution:", distro)

        # Install Docker and Docker Compose based on the distribution
        install_docker_linux(distro)
        install_docker_compose_linux(distro)
    elif os_name == "Darwin":
        # Install Docker Desktop for macOS
        install_docker_mac()
        install_docker_compose_mac()
    elif os_name == "Windows":
        # Install Docker Desktop for Windows
        install_docker_windows()
        install_docker_compose_windows()
    else:
        print("Unsupported operating system:", os_name)
        return

    # Build and run the application using Docker Compose
    print("Building and running the application...")
    try:
        subprocess.run(["docker-compose", "up", "--build", "-d"], check=True)
    except subprocess.CalledProcessError as e:
        print("Error running Docker Compose:", e)
        return

    print("Application is running!")

# ... (previous functions remain the same)

def install_homebrew():
    # Check if Homebrew is installed
    if shutil.which("brew") is None:
        # Install Homebrew
        subprocess.run(["/bin/bash", "-c", "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"], check=True)
        print("Homebrew installed successfully.")
    else:
        print("Homebrew is already installed.")

def install_chocolatey():
    # Check if Chocolatey is installed
    if shutil.which("choco") is None:
        # Install Chocolatey
        subprocess.run(["powershell", "-Command", "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))"], check=True)
        print("Chocolatey installed successfully.")
    else:
        print("Chocolatey is already installed.")

if __name__ == "__main__":
    main()