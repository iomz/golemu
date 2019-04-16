# golemu

[![Build Status](https://travis-ci.org/iomz/golemu.svg?branch=master)](https://travis-ci.org/iomz/golemu)
[![Coverage Status](https://coveralls.io/repos/iomz/golemu/badge.svg?branch=master)](https://coveralls.io/github/iomz/golemu?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/iomz/golemu)](https://goreportcard.com/report/github.com/iomz/golemu)
[![GoDoc](https://godoc.org/github.com/iomz/golemu?status.svg)](http://godoc.org/github.com/iomz/golemu)

A simple LLRP-based RFID reader emulator with [go-llrp](https://github.com/iomz/go-llrp). This emulator was developed as part of [my Master's thesis](https://web.sfc.wide.ad.jp/~iomz/public/mthesis).

Try the [demo](https://github.com/iomz/docker-gosstrak-demo) with docker-compose.

golemu emulates LLRP (Low Level Reader Protocol) communications for RFID inventories via [EPC Gen 2 Class 1 UHF standard](https://www.gs1.org/standards/epc-rfid/uhf-air-interface-protocol) and [ISO/IEC 18000-6 Type C](https://www.iso.org/standard/59644.html).

Please refer to [the original LLRP standard](https://www.gs1.org/standards/epc-rfid/llrp/1-1-0) or [the ISO/IEC equivalent](https://www.iso.org/standard/60833.html).

# Installation

Install [dep](https://github.com/golang/dep) in your system first.

```
$ go get github.com/iomz/golemu
$ cd $GOPATH/src/github.com/iomz/golemu
$ dep ensure && go install .
```

# Synopsis

golemu has 3 commands (`client`, `server`, and `simulator`) to operate in different modes. See the help output for each command.

## LLRP client mode(`golemu client`)

The client mode establishes an LLRP connection with an LLRP server (interrogator). Retry connecting to the server until it becomes online and keep receiving the events until the server closes the connection. 

This command is test-only; which means it only displays the number of events (TAG REPORT DATA parameter) in each received RO ACCESS REPORT message.

## LLRP server mode(`golemu server`)

The server mode first loads tags from a file (`tags.gob` by default) to produce a "virtual inventory of tags." The gob encoded file speeds up the loading process of tags since it is critical for the emulation of hundreds to thousands of tags.

Follow the steps below to generate a gob file; however, this is at the moment *very ugly* and as golemu and [go-llrp](https://github.com/iomz/go-llrp) only takes PC bits and EPC data parameter to represent a tag. I may come up with another way or format to load up tags to golemu in the future.

1. Create a CSV file (e.g., `tags.csv`)

Each line should contain *HEX-encoded PC Bits* and *Binary string of EPC data*.

eg.)
`3000,001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010`

- PC bits: 3000 (indicates a general SGTIN-96 tag)
- EPC Data: 001100000111001000100111011000100111111100101110101001001000000000000000000000000001110001101010 (`307227627f2ea48000001c6a` in HEX)
      
2. Install `gobtags` command

`% go get github.com/iomz/gosstrak/cmd/gobtags`

3. Encode the CSV file to gob file

`% gobtags tags.csv -o tags.gob`

The resulting gob file of the above example (just 1 SGTIN-96 tag) should look like as follows:
```
% hexdump tags.gob
0000000 0a ff 81 06 01 02 ff 84 00 00 00 0a ff 85 06 01
0000010 02 ff 88 00 00 00 2e ff 82 00 2a 03 04 00 02 0a
0000020 ff 85 06 01 02 ff 88 00 00 00 1a ff 86 00 16 05
0000030 06 00 fe 30 00 0f 0a 00 0c 30 72 27 62 7f 2e a4
0000040 80 00 00 1c 6a
0000045
```

(TODO: API documentation for add/delete tags from the virtual tag population)

## Interrogation simulator mode(`golemu simulator`):

The simulator mode also operates as an LLRP server (RFID interrogator) – the only difference is that it iterates through generated gob files in a directory. This mode is intended to simulate a batch of event cycles designed to evaluate specific situations.

One gob file for one event cycle. golemu iterates the file by the file names in ascending order. Suppose there are 3 files (`00.gob`, `01.gob`, and `02.gob` for example) in `/path/to/sim`,

`% golemu simulator -i 2000 /path/to/sim`

repeats generating 3 event cycles with an interval of 2000 ms.

# MTU for LLRP

The PDU option (`-m` or `--pdu`) specifies the maximum allowed PDU and is by default the same as TCP's default MTU (=1500 bytes). golemu adjusts the size of a single RO ACCESS REPORT by limiting the number of TAG REPORT DATA not to exceed the PDU size.

Note that LLRP sometimes doesn't specify the maximum number of parameters in a single message; theoretically, an LLRP packet can grow to the maximum segment size of TCP packets, but this severely affects the overall performance depending on the traffic or the network configurations.

## License

Please cite the following paper.

Mizutani, I., & Mitsugi, J. (2016). A Multicode and Portable RFID Tag Events Emulator for RFID Information System. Proceedings of the 6th International Conference on the Internet of Things  - IoT’16, 187–188. https://doi.org/10.1145/2991561.2998470

See the LICENSE file.

## Author

Iori Mizutani (iomz)
