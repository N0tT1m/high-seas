import os
import subprocess
import platform

def run_command(command):
    process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, error = process.communicate()
    if process.returncode != 0:
        print(f"Error executing command: {command}")
        print(f"Error: {error.decode('utf-8')}")
    else:
        print(f"Command executed successfully: {command}")

# Determine the operating system
os_name = platform.system()

# Install npm and Node.js
if os_name == "Linux":
    run_command("sudo apt-get update")
    run_command("sudo apt-get install -y nodejs npm")
elif os_name == "Darwin":  # macOS
    run_command("brew update")
    run_command("brew install node")
elif os_name == "Windows":
    run_command("winget install OpenJS.NodeJS")
else:
    print("Unsupported operating system.")
    exit(1)

# Install Golang
if os_name == "Linux":
    run_command("sudo apt-get install -y golang")
elif os_name == "Darwin":  # macOS
    run_command("brew install go")
elif os_name == "Windows":
    run_command("winget install GoLang.Go")
else:
    print("Unsupported operating system.")
    exit(1)

# Install Docker Desktop
if os_name == "Linux":
    run_command("sudo apt-get install -y docker.io")
elif os_name == "Darwin":  # macOS
    run_command("brew install --cask docker")
elif os_name == "Windows":
    run_command("winget install Docker.DockerDesktop")
else:
    print("Unsupported operating system.")
    exit(1)

# Create the './plex/config.py' file
os.makedirs("./plex", exist_ok=True)
with open("./plex/config.py", "w") as file:
    file.write("""HOST="192.168.1.1"
USER="root"
PASSWD="ThisIsAPassword"
DB="highseas"
IP="192.168.1.1"
PORT="32400"
""")

# Create the './web/nginx.conf' file
os.makedirs("./web", exist_ok=True)
with open("./web/nginx.conf", "w") as file:
    file.write("""# the events block is required
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
""")

# Create the environment files
os.makedirs("./web/src/app/environments", exist_ok=True)
environment_files = ["environment.prod.ts", "environment.ts", "environment.deployment.ts"]
for file_name in environment_files:
    with open(f"./web/src/app/environments/{file_name}", "w") as file:
        file.write("""export const environment = {
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
""")