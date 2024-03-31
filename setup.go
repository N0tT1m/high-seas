package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

func runCommand(command string, wg *sync.WaitGroup) {
	defer wg.Done()

	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %s\n", command)
		fmt.Printf("Error: %s\n", string(output))
	} else {
		fmt.Printf("Command executed successfully: %s\n", command)
	}
}

func main() {
	// Determine the operating system
	osName := runtime.GOOS

	var wg sync.WaitGroup

	// Install npm and Node.js
	wg.Add(1)
	go func() {
		defer wg.Done()
		if osName == "linux" {
			runCommand("sudo apt-get update", &wg)
			runCommand("sudo apt-get install -y nodejs npm", &wg)
		} else if osName == "darwin" {
			runCommand("brew update", &wg)
			runCommand("brew install node", &wg)
		} else if osName == "windows" {
			runCommand("winget install OpenJS.NodeJS", &wg)
		} else {
			fmt.Println("Unsupported operating system.")
			os.Exit(1)
		}
	}()

	// Install Golang
	wg.Add(1)
	go func() {
		defer wg.Done()
		if osName == "linux" {
			runCommand("sudo apt-get install -y golang", &wg)
		} else if osName == "darwin" {
			runCommand("brew install go", &wg)
		} else if osName == "windows" {
			runCommand("winget install GoLang.Go", &wg)
		} else {
			fmt.Println("Unsupported operating system.")
			os.Exit(1)
		}
	}()

	// Install Docker Desktop
	wg.Add(1)
	go func() {
		defer wg.Done()
		if osName == "linux" {
			runCommand("sudo apt-get install -y docker.io", &wg)
		} else if osName == "darwin" {
			runCommand("brew install --cask docker", &wg)
		} else if osName == "windows" {
			runCommand("winget install Docker.DockerDesktop", &wg)
		} else {
			fmt.Println("Unsupported operating system.")
			os.Exit(1)
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

	wg.Wait()
}
