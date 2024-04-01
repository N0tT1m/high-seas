# High Seas


**The High Seas app is designed to allow you to look for new or old shows / movies and allow you to download them.**


## NOTE: This is purely for educational purposes.


### Environment for High Seas typescript file format

#### Location to this file is: ```./web/src/app/environments/environment.ts```

```
export const environment = {
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
  
```

### Ngnix Config File

#### Location to this file is: ```./web/nginx.conf```

```
# the events block is required
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

```

## Install script

Location: `./install-scripts`

The scripts come in two forms:

A Go script:
  - Setup.go

A Python script:
  - Setup.py

The Go script works, currently the Python script isn't refactored.

## High Seas Frontend

### Running the frontend of High Seas

#### Installing NodeJS on Windows

Download NodeJS from: https://nodejs.org/dist/v20.12.0/node-v20.12.0-x64.msi

#### Running the app locally

```
cd web
npm install
npm run start-local
```

### Building the frontend of High Seas

To run the frontend of High Seas you can run the command:
```
cd ./web
dockerbuild -t high-seas-frontend .
docker run -d -p 6969:6969 high-seas-frontend
```


## High Seas Backend

### Building the backend of High Seas

To run the backend of High Seas you can run the command:
```
dockerbuild -t high-seas-backend .
docker run -d -p 8782:8782 high-seas-backend
```

##### Default file name **.env** in the root directory

```
DB_USER=DB_USER
DB_PASSWORD=DB_PASSWORD
DB_IP=DB_IP
DB_PORT=DB_PORT
DELUGE_IP=DELUGE_IP
DELUGE_PORT=DELUGE_PORT
DELUGE_USER=DELUGE_USER
DELUGE_PASSWORD=DELUGE_PASSWORD
JACKETT_IP=JACKETT_IP_HERE
JACKETT_PORT=JACKETT_PORT_HERE
JACKETT_API_KEY=YOUR_KEY_HERE
```

## Plex Python Backend

### Config File Example

##### Default file name: **config.py**

```
HOST="192.168.1.1"
USER="root"
PASSWD="ThisIsAPassword"
DB="highseas"
IP="192.168.1.1"
PORT="32400"
```

## Docker Compose

### Running Docker Compose Yaml

#### Linux

To run docker-compose on Linux you need to run the script: ```start-dedicated.sh```