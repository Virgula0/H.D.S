# Distributed Programming University Project

<img src='/docs/images/logo.png' style='zoom: 20%; border-radius: 50%; margin-right: 3%' align=left />

<font size='10'><strong>H.D.S</strong></font><br>

1<sup>st</sup> January 2024

Student: <font color='orange'>Angelo Rosa</font>

<br><br>

# The project

<img src='/docs/images/project-structure.png' style='zoom: 100%;  border: 2px solid #ddd;'  alt="missing"/>

Project structure

```
├── client
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── internal
│   │   ├── constants
│   │   │   └── constants.go
│   │   ├── entities
│   │   │   ├── auth_request.go
│   │   │   └── handshake.go
│   │   ├── environment
│   │   │   └── init.go
│   │   ├── grpcclient
│   │   │   ├── communication.go
│   │   │   └── init.go
│   │   ├── gui
│   │   │   ├── login_window.go
│   │   │   └── process_window.go
│   │   ├── hcxtools
│   │   │   └── hcxpcapngtool.go
│   │   ├── mygocat
│   │   │   ├── gocat.go
│   │   │   └── task_handler.go
│   │   ├── resources
│   │   │   └── fonts
│   │   │       ├── Roboto-BlackItalic.ttf
│   │   │       ├── Roboto-Black.ttf
│   │   │       ├── Roboto-BoldItalic.ttf
│   │   │       ├── Roboto-Bold.ttf
│   │   │       ├── Roboto-Italic.ttf
│   │   │       ├── Roboto-LightItalic.ttf
│   │   │       ├── Roboto-Light.ttf
│   │   │       ├── Roboto-MediumItalic.ttf
│   │   │       ├── Roboto-Medium.ttf
│   │   │       ├── Roboto-Regular.ttf
│   │   │       ├── Roboto-ThinItalic.ttf
│   │   │       └── Roboto-Thin.ttf
│   │   └── utils
│   │       └── utils.go
│   ├── main.go
│   ├── Makefile
│   └── wordlists
├── database
│   ├── Dockerfile
│   ├── initialize.sql
│   └── my.cnf
├── docker-compose.yaml
├── externals
│   ├── gocat
│   ├── hashcat
│   └── hcxtools
├── LICENSE
├── proto-definitions
│   └── hds
│       ├── hds.proto
│       ├── hds_request.proto
│       └── hds_response.proto
├── proto.sh
├── raspberry-pi
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── handshakes
│   │   └── test.pcap
│   ├── internal
│   │   ├── authapi
│   │   │   └── authenticate.go
│   │   ├── cmd
│   │   │   └── command_parser.go
│   │   ├── constants
│   │   │   └── constants.go
│   │   ├── deamon
│   │   │   ├── communication.go
│   │   │   ├── environment.go
│   │   │   └── init.go
│   │   ├── entities
│   │   │   ├── api_entities.go
│   │   │   └── handshake.go
│   │   ├── utils
│   │   │   └── utils.go
│   │   ├── wifi
│   │   │   └── wifi.go
│   │   └── wpaparser
│   │       ├── getwpa.go
│   │       └── parser.go
│   ├── main.go
│   └── Makefile
├── README.md
└── server
    ├── backend
    │   ├── cmd
    │   │   └── main.go
    │   └── internal
    │       ├── constants
    │       │   └── constants.go
    │       ├── errors
    │       │   └── errors.go
    │       ├── grpcserver
    │       │   ├── commands.go
    │       │   ├── common_grpc_test.go
    │       │   ├── controllers.go
    │       │   ├── grpc_test.go
    │       │   ├── init.go
    │       │   └── options.go
    │       ├── infrastructure
    │       │   └── database.go
    │       ├── raspberrypi
    │       │   ├── common_raspberrypi_test.go
    │       │   ├── components.go
    │       │   ├── init.go
    │       │   ├── raspberrypi_test.go
    │       │   └── tcp_server.go
    │       ├── repository
    │       │   └── repository.go
    │       ├── response
    │       │   └── response.go
    │       ├── restapi
    │       │   ├── authenticate
    │       │   │   ├── handler_anonymous.go
    │       │   │   └── handler_user.go
    │       │   ├── client
    │       │   │   └── handler_user.go
    │       │   ├── handlers.go
    │       │   ├── handshake
    │       │   │   └── handler_user.go
    │       │   ├── logout
    │       │   │   └── handler_user.go
    │       │   ├── middlewares
    │       │   │   ├── auth_middlware.go
    │       │   │   ├── common_middleware.go
    │       │   │   └── log_requests.go
    │       │   ├── raspberrypi
    │       │   │   └── handler_user.go
    │       │   ├── register
    │       │   │   └── anonymous_handler.go
    │       │   └── routes.go
    │       ├── seed
    │       │   └── seed_api.go
    │       ├── testsuite
    │       │   ├── auth_api.go
    │       │   ├── setup_grpc.go
    │       │   └── tcp_ip.go
    │       ├── usecase
    │       │   └── usecase.go
    │       └── utils
    │           ├── utils.go
    │           └── validator.go
    ├── Dockerfile
    ├── entities
    │   ├── client.go
    │   ├── handshake.go
    │   ├── raspberry_pi.go
    │   ├── role.go
    │   ├── uniform_response.go
    │   └── user.go
    ├── frontend
    │   ├── cmd
    │   │   ├── custom.go
    │   │   └── main.go
    │   ├── internal
    │   │   ├── constants
    │   │   │   └── constants.go
    │   │   ├── errors
    │   │   │   └── errors.go
    │   │   ├── middlewares
    │   │   │   ├── auth_middleware.go
    │   │   │   ├── cookie_middleware.go
    │   │   │   └── log_requests.go
    │   │   ├── pages
    │   │   │   ├── clients
    │   │   │   │   └── clients.go
    │   │   │   ├── handshakes
    │   │   │   │   └── handshake.go
    │   │   │   ├── login
    │   │   │   │   └── login.go
    │   │   │   ├── logout
    │   │   │   │   └── logout.go
    │   │   │   ├── pages.go
    │   │   │   ├── raspberrypi
    │   │   │   │   └── raspberrypi.go
    │   │   │   ├── register
    │   │   │   │   └── register.go
    │   │   │   ├── routes.go
    │   │   │   └── welcome
    │   │   │       └── welcome.go
    │   │   ├── repository
    │   │   │   └── repository.go
    │   │   ├── response
    │   │   │   └── response.go
    │   │   ├── usecase
    │   │   │   └── usecase.go
    │   │   └── utils
    │   │       ├── utils.go
    │   │       └── validator.go
    │   ├── static
    │   │   ├── images
    │   │   │   └── logo.png
    │   │   ├── scripts
    │   │   │   ├── bootstrap.min.js
    │   │   │   ├── dashboard.js
    │   │   │   ├── github-stats.js
    │   │   │   ├── jquery-3.3.1.min.js
    │   │   │   ├── popper.min.js
    │   │   │   └── theme-toggle.js
    │   │   ├── static.go
    │   │   └── styles
    │   │       ├── bootstrap-4.3.1.min.css
    │   │       ├── custom.css
    │   │       └── main.css
    │   └── views
    │       ├── clients.html
    │       ├── handshake.html
    │       ├── login.html
    │       ├── raspberrypi.html
    │       ├── register.html
    │       ├── views.go
    │       └── welcome.html
    ├── go.mod
    ├── go.sum
    ├── main.go
    └── Makefile

79 directories, 143 files
```

## Run the application (in test mode) with docker

Since `client` within docker runs a GUI using `raylib`, we need to forward the desktop environment to docker.
So, a utility called `xhost` is needed (`xorg-xhost` package on arch).

Then run 

```bash
export DISPLAY=:0.0 && \
xhost +local:docker && \
docker compose up --build
```

# Deamon

[Setup](raspberry-pi/README.md)

# Client

[Setup](client/README.md)