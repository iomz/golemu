gologir
==

[![Build Status](https://travis-ci.org/iomz/gologir.svg?branch=master)](https://travis-ci.org/iomz/gologir)
[![Coverage Status](https://coveralls.io/repos/iomz/gologir/badge.svg?branch=master)](https://coveralls.io/github/iomz/gologir?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/iomz/gologir)](https://goreportcard.com/report/github.com/iomz/gologir)
[![GoDoc](https://godoc.org/github.com/iomz/gologir?status.svg)](http://godoc.org/github.com/iomz/gologir)

A simple LLRP-based logical reader mock for RFID Tags using [go-llrp](https://github.com/iomz/go-llrp).

For what?
--

You need to confirm the tag stream events in scenarios, but not really thrilled to use a heavily armed simulator like RIFIDI Edge suite.

Install & Synopsis
--

```
$ go get github.com/iomz/gologir
$ gologir --help

usage: gologir [<flags>] <command> [<args> ...]

A mock LLRP-based logical reader for RFID Tags.

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
  -v, --verbose                Enable verbose mode.
  -m, --initialMessageID=1000  The initial messageID to start from.
  -k, --initialKeepaliveID=80000
                               The initial keepaliveID to start from.
  -p, --port=5084              LLRP listening port.
  -i, --ip=0.0.0.0             LLRP listening address.
      --version                Show application version.

Commands:
  help [<command>...]
    Show help.

  server [<flags>]
    Run as a tag stream server.

  client
    Run as a client mode.
```

Run as a server, listen 5084 port for LLRP incoming connection, 8080 port for websocket UI

```
$ gologir server
Access http://localhost:8080 for Web GUI
```

Links
--

https://gin-gonic.github.io/gin/

https://github.com/fatih/structs

http://gopkg.in/alecthomas/kingpin.v2

https://godoc.org/golang.org/x/net/websocket

Author
--

Iori Mizutani (iomz)

License
--

```
The MIT License (MIT)
Copyright © 2016 Iori MIZUTANI <iori.mizutani@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the “Software”), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
