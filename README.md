# golemu

[![Build Status](https://travis-ci.org/iomz/golemu.svg?branch=master)](https://travis-ci.org/iomz/golemu)
[![Coverage Status](https://coveralls.io/repos/iomz/golemu/badge.svg?branch=master)](https://coveralls.io/github/iomz/golemu?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/iomz/golemu)](https://goreportcard.com/report/github.com/iomz/golemu)
[![GoDoc](https://godoc.org/github.com/iomz/golemu?status.svg)](http://godoc.org/github.com/iomz/golemu)

A simple LLRP-based RFID reader emulator with [go-llrp](https://github.com/iomz/go-llrp).
Try the [demo](https://github.com/iomz/docker-gosstrak-demo) with docker-compose.

## For what?

You need to confirm the tag stream events in scenarios, but not really thrilled to use a heavily armed simulator like RIFIDI Edge suite.

## Install & Synopsis

Install [dep](https://github.com/golang/dep) in your system first.

```
$ go get github.com/iomz/golemu
$ cd $GOPATH/src/github.com/iomz/golemu
$ dep ensure
$ golemu --help

usage: golemu [<flags>] <command> [<args> ...]

A mock LLRP-based logical reader emulator for RFID Tags.

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
  -v, --debug                  Enable debug mode.
      --initialMessageID=1000  The initial messageID to start from.
      --initialKeepaliveID=80000
                               The initial keepaliveID to start from.
  -a, --ip=0.0.0.0             LLRP listening address.
  -k, --keepalive=0            LLRP Keepalive interval.
  -p, --port=5084              LLRP listening port.
  -m, --pdu=1500               The maximum size of LLRP PDU.
  -i, --reportInterval=10000   The interval of ROAccessReport in ms. Pseudo ROReport spec option.
      --version                Show application version.

Commands:
  help [<command>...]
    Show help.

  server [<flags>]
    Run as an LLRP tag stream server.

  client
    Run as an LLRP client.

  simulate <simulationDir>
    Run in the simulator mode.


```

## License

MIT

## Author

Iori Mizutani (iomz)
