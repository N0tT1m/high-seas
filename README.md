# High Seas


**The High Seas app is designed to allow you to look for new or old shows / movies and allow you to download them.**


### NOTE: This is purely for educational purposes.


## High Seas Frontend

### Building the front of High Seas

To run the frontend of High Seas you can run the command ```dockerbuild -t high-seas-frontend .```
Then ````docker run -d -p 8889:8889 high-seas-frontend```


### Environment for High Seas typescript file format

#### Location to this file is ```./web/src/app/environments/```

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


## High Seas Backend

### Building the backend of High Seas

To run the backend of High Seas you can run the command ```dockerbuild -t high-seas-backend .```
Then ```docker run -d -p 8782:8782 high-seas-backend```

## Docker Compose

### Running Docker Compose Yaml

#### Linux

To run docker-compose on Linux you need to run the script ```start-dedicated.sh```

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