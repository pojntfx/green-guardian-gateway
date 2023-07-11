# GreenGuardian Gateway

![Logo](./docs/logo-readme.png)

AWS IoT Gateway for GreenGuardian, a uni project for the HdM IoT course.

[![hydrun CI](https://github.com/pojntfx/green-guardian-gateway/actions/workflows/hydrun.yaml/badge.svg)](https://github.com/pojntfx/green-guardian-gateway/actions/workflows/hydrun.yaml)
[![Docker CI](https://github.com/pojntfx/green-guardian-gateway/actions/workflows/docker.yaml/badge.svg)](https://github.com/pojntfx/green-guardian-gateway/actions/workflows/docker.yaml)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.20-61CFDD.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/pojntfx/green-guardian-gateway.svg)](https://pkg.go.dev/github.com/pojntfx/green-guardian-gateway)
[![Matrix](https://img.shields.io/matrix/green-guardian-gateway:matrix.org)](https://matrix.to/#/#green-guardian-gateway:matrix.org?via=matrix.org)
[![Binary Downloads](https://img.shields.io/github/downloads/pojntfx/green-guardian-gateway/total?label=binary%20downloads)](https://github.com/pojntfx/green-guardian-gateway/releases)

## Overview

The GreenGuardian Gateway connects your sensors and actuators to the GreenGuardian Cloud Backend.

It enables you too ...

- **Map sensors and actuators**: By connecting hubs with sensors and actuators to a central gateway, you can model a real-world topology of rooms and plants using the network.
- **Export moisture and temperature data**: The gateway forwards the periodically measured moisture and temperature data from the hub to the cloud, allowing it to be processed there.
- **Remotely control the sensors**: When the GreenGuardian Cloud Backend decides to take action and turn on a sprinkler or fan, the gateway forwards this request to the correct hub, which then forwards it to the correct actuator.

## Installation

### Containerized

You can get the OCI images like so:

```shell
$ podman pull ghcr.io/pojntfx/green-guardian-gateway
$ podman pull ghcr.io/pojntfx/green-guardian-hub
```

### Natively

Static binaries are available on [GitHub releases](https://github.com/pojntfx/green-guardian-gateway/releases).

On Linux, you can install them like so:

```shell
$ curl -L -o /tmp/green-guardian-gateway "https://github.com/pojntfx/green-guardian-gateway/releases/latest/download/green-guardian-gateway.linux-$(uname -m)"
$ curl -L -o /tmp/green-guardian-hub "https://github.com/pojntfx/green-guardian-gateway/releases/latest/download/green-guardian-hub.linux-$(uname -m)"
$ sudo install /tmp/green-guardian-gateway /usr/local/bin
$ sudo install /tmp/green-guardian-hub /usr/local/bin
```

On Windows, the following should work (using PowerShell as administrator):

```shell
PS> Invoke-WebRequest https://github.com/pojntfx/green-guardian-gateway/releases/latest/download/green-guardian-gateway.windows-x86_64.exe -OutFile \Windows\System32\green-guardian-gateway.exe
PS> Invoke-WebRequest https://github.com/pojntfx/green-guardian-gateway/releases/latest/download/green-guardian-hub.windows-x86_64.exe -OutFile \Windows\System32\green-guardian-hub.exe
```

You can find binaries for more architectures on [GitHub releases](https://github.com/pojntfx/green-guardian-gateway/releases).

## Reference

### Command Line Arguments

#### Gateway

```shell
$ green-guardian-gateway --help
Usage of green-guardian-gateway:
  -aws-ca string
        AWS mTLS CA (default "/home/pojntfx/Projects/green-guardian-gateway/crypto/ca.pem")
  -aws-cert string
        AWS mTLS certificate (default "/home/pojntfx/Projects/green-guardian-gateway/crypto/cert.pem")
  -aws-key string
        AWS mTLS secret key (default "/home/pojntfx/Projects/green-guardian-gateway/crypto/key.pem")
  -endpoint string
        AWS MQTT endpoint to connect to (default "ssl://ad218s2flbk57-ats.iot.eu-central-1.amazonaws.com:8883")
  -laddr string
        Listen address (default ":1337")
  -thing-name string
        Thing name (for topic to publish too; invalid thing names are denied using the ) (default "DEVICE-Device_1")
  -verbose
        Whether to enable verbose logging
```

#### Hub

```shell
$ green-guardian-hub --help
Usage of green-guardian-hub:
  -baud int
        Baudrate to use to communicate with sensors and actuators (default 115200)
  -default-moisture int
        The default expected moisture (default 30)
  -default-temperature int
        The default expected temperature (default 25)
  -fans string
        JSON description in the format { roomID: devicePath } (default "{\"1\": \"/dev/ttyACM0\"}")
  -measure-interval duration
        Amount of time after which a new measurement is taken (default 1s)
  -measure-timeout duration
        Amount of time after which it is assumed that a measurement has failed (default 1s)
  -mock int
        If set to >1, mock temperature and moisture using buttons, sending the default value +- the value of this flag
  -moisture-sensors string
        JSON description in the format { roomID: devicePath } (default "{\"1\": \"/dev/ttyACM0\"}")
  -raddr string
        Remote address (default "localhost:1337")
  -sprinklers string
        JSON description in the format { plantID: devicePath } (default "{\"1\": \"/dev/ttyACM0\"}")
  -temperature-sensors string
        JSON description in the format { roomID: devicePath } (default "{\"1\": \"/dev/ttyACM0\"}")
  -verbose
        Whether to enable verbose logging
```

### Environment Variables

You can set some flags using environment variables. For more info, see the [docker-compose file](./docker-compose.yaml).

## Acknowledgements

- [eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) provides the MQTT client library.
- [golang/mock](https://github.com/golang/mock) provides the mocking library.
- [pojntfx/dudirekta](https://github.com/pojntfx/dudirekta) provides the RPC framework used for communicating between the gateway and the hub.

## Contributing

To contribute, please use the [GitHub flow](https://guides.github.com/introduction/flow/) and follow our [Code of Conduct](./CODE_OF_CONDUCT.md).

To build and start a development version of the GreenGuardian Gateway locally, run the following:

```shell
$ git clone https://github.com/pojntfx/green-guardian-gateway.git
$ cd green-guardian-gateway
$ make depend
$ make && sudo make install
$ green-guardian-gateway # Adjust flags for your own AWS configuration
# In another terminal
$ sudo green-guardian-hub # Adjust flags for your own USB configuration
```

For more information, esp. on how to set up your AWS infrastructure, see [docs/demo.md](./docs/demo.md).

Have any questions or need help? Chat with us [on Matrix](https://matrix.to/#/#green-guardian-gateway:matrix.org?via=matrix.org)!

## License

GreenGuardian Gateway (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: AGPL-3.0
