# thorium-go

A game server cluster management tool using HTTP/REST.

Can run as a one-node or multi-node cluster. Some knowledge of network management and configuration is required to setup and run. I.e. setting up machines, exposing addresses and ports, etc. 

### Task Tracking

This project is under development. To see the roadmap please view the public Trello Board: https://trello.com/b/Eb2I4WiC/thorium-go-development

### Setup Requirements

- Golang (1.5+ recommended)
- Docker (1.8+ recommended)
- Docker-Compose (1.6+ recommended)


### Setup Instructions

##### Download the project repository 

This project is *go gettable*, which means that the ```go get``` command will download the project and all dependencies.

```
go get github.com/jaybennett89/thorium-go
```

You may find that some dependecies are missing when you try to build. In this case, you  have to fetch the missing dependecy manually. I had inconsistent results with the above command on whether it would properly fetch any of the dependencies for this project. However, all dependences are *go gettable* manually.

```
go get gopkg.in/redis.v3
go get github.com/dgrijalva/jwt-go
go get github.com/go-martini/martini
go get github.com/jaybennett89/intmath/intgr
go get github.com/lib/pq
```

##### Build and Run the Master Node

The Master node runs as a set of Docker containers containing Postgres, Redis, and the main Master process. To build the master node we must first build the Master binary and then create a Docker image. I have included a handy Makefile in the ```/cmd/master-server``` directory.

```
cd /thorium-go/cmd/masterserver
make image
```

The Docker image only needs to be created once. To skip rebuilding the Docker image, a second Makefile command is provided.

```
make build
```

Generate the RSA keys used to secure your JSON Web Tokens.

```
cd /thorium-go/keys
openssl genrsa -out app.rsa 2048
openssl rsa -in app.rsa -pubout > app.rsa.pub
```

Finally, we are ready to launch the Master node.

```
cd /thorium-go
docker-compose up -d
```

This will launch the service. 

##### Monitoring the Master Service

You can check the status of the service using the Docker client.

```
$ docker ps

CONTAINER ID        IMAGE                     COMMAND                  CREATED             STATUS              PORTS                    NAMES
3cfda9d9eaad        thoriumgo_master-server   "./master-server"        23 hours ago        Up 23 hours         0.0.0.0:6960->6960/tcp   thoriumgo_master-server_1
c1a5b6a6fd24        library/postgres          "/docker-entrypoint.s"   23 hours ago        Up 23 hours         5432/tcp                 thoriumgo_db_1
ca556ec1d20a        redis                     "/entrypoint.sh redis"   23 hours ago        Up 23 hours         6379/tcp                 thoriumgo_cache_1
```


You can monitor it by watching the docker log.

```
$ docker logs thoriumgo_master-server_1

2016/04/20 03:59:52 opening app.rsa keys
2016/04/20 03:59:52 testing postgres connection
2016/04/20 03:59:52 dial tcp 172.17.0.3:5432: getsockopt: connection refused
2016/04/20 03:59:52 testing redis connection
2016/04/20 03:59:52 thordb initialization complete
[martini] listening on :6960 (development)
```

##### Build and Run A Host Node

A **Host** is the process that manages one or more  **Game Server** processes on a physical machine.

The **Host** process runs as a standalone binary, unlike the **Master** which uses Docker containers. This due to issues with runtime exposing of **Game Server** ports inside a container. 

There are two ways to run the Host node. The first is with the ```go run``` command which will automatically build and run Go source code in the current directory. It is most useful in the development or testing stages when you have to restart the Host node often.

```
cd /thorium-go/cmd/host-server
go run host-server.go
```

The second method is to build and run the Host binary manually. A much better option for long uptimes.

```
go build host-server.go
./host-server > host-server.log
```

#### Testing the Cluster

The project contains an example client, game server, and automated integration test suite. The automated tests are intended to work with the ```example-gameserver```, a mock Game Server written as a Go REST API. This is not intended to be used as a real **Game Server** and should be replaced with your own code (explained further down below).

A Makefile is provided to build the example-gameserver binary and install it in the ```/host-server/bin``` directory.

```
cd /thorium-go/cmd/example-gameserver
make install
```

You can now execute the automated test suite. 

```
cd /thorium-go/client
go test
```

All tests should pass if your cluster is setup correctly!

##### Restarting for Production

Please note that the test suite works against a running cluster and creates records in the database; therefore, it is recommended that you kill and restart the **Master** node after testing.

```
# restarting the Master node
docker-compose kill
docker-compose rm | yes
docker-compose up -d
```

### Getting Started With Your Project

This project is intended to work with any Game Server and Game Client engine (i.e. Unity, UE4, etc. all work). There are only a few required steps for your **Game Server** to implement in order to register correctly with the server. You may use some or all of the features provided by the Master API, such as:

- Accounts (login, register)
- Games (get list, create, join)
- Characters (create, update)

##### Configuring the Host Node

The Host node needs to know what file to use as the Game Server application. This can be changed in the ```host.config``` file found in ```/thorium-go/cmd/host-server```.

```
{
    "GameserverBinaryPath" : "bin/$your_game_server"
}
```

Place your **Game Server** in the ```/host-server/bin``` directory and modify the property in the config file to point to it.

It is recommended to restart the Host server upon changing the host.config.

##### Implementing Your Own Game Server and Client

For tips on implementing a new game server and client that uses the *thorium-go* service, see the reference implementation and test scripts in ```/client/client.go``` directory for demos of different use cases.

The ```example-gameserver``` program outlines how to create a **Game Server** that properly registers itself with the service. A **Game Server** should try to talk to the **Host** service on ```localhost```, instead of communicating with the **Master**.
