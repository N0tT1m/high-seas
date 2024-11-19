package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// Configuration structs for different components
type FrontendConfig struct {
	TMDBToken     string
	GolangAPIPort string
	GolangAPIIP   string
	Transport     string
}

type BackendConfig struct {
	DBUser         string
	DBPassword     string
	DBIP           string
	DBPort         string
	DelugeIP       string
	DelugePort     string
	DelugeUser     string
	DelugePassword string
	JackettIP      string
	JackettPort    string
	JackettAPIKey  string
}

type PlexConfig struct {
	Host     string
	User     string
	Password string
	Database string
	IP       string
	Port     string
}

// promptUser helper function to get user input
func promptUser(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// createFrontendEnvironment creates the TypeScript environment file
func createFrontendEnvironment(config FrontendConfig) error {
	envContent := `export const environment = {
    production: true,
    baseUrl: 'http://www.example.com:8080',
    envVar: {
      authorization: "{{.TMDBToken}}",
      port: "{{.GolangAPIPort}}",
      ip: "{{.GolangAPIIP}}",
      transport: "{{.Transport}}",
    },
  };`

	// Ensure directory exists
	err := os.MkdirAll("./web/src/app/environments", 0755)
	if err != nil {
		return err
	}

	// Create and write to file
	file, err := os.Create("./web/src/app/environments/environment.ts")
	if err != nil {
		return err
	}
	defer file.Close()

	tmpl, err := template.New("environment").Parse(envContent)
	if err != nil {
		return err
	}

	return tmpl.Execute(file, config)
}

// createBackendEnv creates the .env file for the backend
func createBackendEnv(config BackendConfig) error {
	envContent := fmt.Sprintf(`DB_USER=%s
DB_PASSWORD=%s
DB_IP=%s
DB_PORT=%s
DELUGE_IP=%s
DELUGE_PORT=%s
DELUGE_USER=%s
DELUGE_PASSWORD=%s
JACKETT_IP=%s
JACKETT_PORT=%s
JACKETT_API_KEY=%s`,
		config.DBUser,
		config.DBPassword,
		config.DBIP,
		config.DBPort,
		config.DelugeIP,
		config.DelugePort,
		config.DelugeUser,
		config.DelugePassword,
		config.JackettIP,
		config.JackettPort,
		config.JackettAPIKey,
	)

	return os.WriteFile(".env", []byte(envContent), 0644)
}

// createPlexConfig creates the Plex backend config file
func createPlexConfig(config PlexConfig) error {
	configContent := fmt.Sprintf(`HOST="%s"
USER="%s"
PASSWD="%s"
DB="%s"
IP="%s"
PORT="%s"`,
		config.Host,
		config.User,
		config.Password,
		config.Database,
		config.IP,
		config.Port,
	)

	return os.WriteFile("config.py", []byte(configContent), 0644)
}

// buildDockerContainers builds and starts the Docker containers
func buildDockerContainers() error {
	// Build frontend
	fmt.Println("Building frontend container...")
	cmdFront := exec.Command("docker", "build", "-t", "high-seas-frontend", "./web")
	cmdFront.Stdout = os.Stdout
	cmdFront.Stderr = os.Stderr
	if err := cmdFront.Run(); err != nil {
		return fmt.Errorf("frontend build failed: %v", err)
	}

	// Build backend
	fmt.Println("Building backend container...")
	cmdBack := exec.Command("docker", "build", "-t", "high-seas-backend", ".")
	cmdBack.Stdout = os.Stdout
	cmdBack.Stderr = os.Stderr
	if err := cmdBack.Run(); err != nil {
		return fmt.Errorf("backend build failed: %v", err)
	}

	// Start containers using docker-compose
	fmt.Println("Starting containers with docker-compose...")
	cmdCompose := exec.Command("docker-compose", "up", "-d")
	cmdCompose.Stdout = os.Stdout
	cmdCompose.Stderr = os.Stderr
	if err := cmdCompose.Run(); err != nil {
		return fmt.Errorf("docker-compose failed: %v", err)
	}

	return nil
}

func main() {
	fmt.Println("High Seas Setup Script")
	fmt.Println("=====================")

	// Get Frontend Configuration
	frontendConfig := FrontendConfig{
		TMDBToken:     promptUser("Enter TMDB API Bearer Token: "),
		GolangAPIPort: promptUser("Enter Golang API Port: "),
		GolangAPIIP:   promptUser("Enter Golang API IP: "),
		Transport:     promptUser("Enter Transport (HTTP/HTTPS): "),
	}

	// Get Backend Configuration
	backendConfig := BackendConfig{
		DBUser:         promptUser("Enter Database User: "),
		DBPassword:     promptUser("Enter Database Password: "),
		DBIP:           promptUser("Enter Database IP: "),
		DBPort:         promptUser("Enter Database Port: "),
		DelugeIP:       promptUser("Enter Deluge IP: "),
		DelugePort:     promptUser("Enter Deluge Port: "),
		DelugeUser:     promptUser("Enter Deluge User: "),
		DelugePassword: promptUser("Enter Deluge Password: "),
		JackettIP:      promptUser("Enter Jackett IP: "),
		JackettPort:    promptUser("Enter Jackett Port: "),
		JackettAPIKey:  promptUser("Enter Jackett API Key: "),
	}

	// Get Plex Configuration
	plexConfig := PlexConfig{
		Host:     promptUser("Enter Plex Host: "),
		User:     promptUser("Enter Plex User: "),
		Password: promptUser("Enter Plex Password: "),
		Database: promptUser("Enter Plex Database Name: "),
		IP:       promptUser("Enter Plex IP: "),
		Port:     promptUser("Enter Plex Port: "),
	}

	// Create configuration files
	fmt.Println("\nCreating configuration files...")

	if err := createFrontendEnvironment(frontendConfig); err != nil {
		fmt.Printf("Error creating frontend environment: %v\n", err)
		return
	}

	if err := createBackendEnv(backendConfig); err != nil {
		fmt.Printf("Error creating backend .env: %v\n", err)
		return
	}

	if err := createPlexConfig(plexConfig); err != nil {
		fmt.Printf("Error creating Plex config: %v\n", err)
		return
	}

	// Build and start Docker containers
	fmt.Println("\nBuilding and starting Docker containers...")
	if err := buildDockerContainers(); err != nil {
		fmt.Printf("Error with Docker operations: %v\n", err)
		return
	}

	fmt.Println("\nSetup completed successfully!")
	fmt.Println("You can access the application at:")
	fmt.Printf("Frontend: http://localhost:6969\n")
	fmt.Printf("Backend: http://localhost:8782\n")
}
