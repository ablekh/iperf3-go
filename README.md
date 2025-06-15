# iperf3-go

A complete Go implementation of iperf3 with both client and server functionality, fully compatible with the standard iperf3 tools.

## Features

- **Complete iperf3 implementation** - Both client and server modes
- **Multiple protocol support** - TCP, UDP, and SCTP (Linux only)
- **TCP performance testing** with accurate measurements
- **UDP performance testing** with bandwidth control and packet rate limiting
- **SCTP performance testing** with multi-stream support (Linux only)
- **Full compatibility** with standard iperf3 clients and servers
- **JSON output format** matching iperf3 specification
- **Real-time interval reporting** during tests
- **Multiple client support** for server mode
- **Cross-platform** support (Windows, Linux, macOS)

## Prerequisites

You need to have Go installed on your system. If Go is not installed:

### Installing Go on Windows:
1. Download Go from https://golang.org/dl/
2. Run the installer and follow the instructions
3. Verify installation: `go version`

### Installing Go on Linux/macOS:
```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install golang-go

# CentOS/RHEL
sudo yum install golang

# macOS with Homebrew
brew install go
```

## Installation

1. Clone or download this project
2. Navigate to the project directory
3. Build the server:

```bash
go build -o iperf3-go main.go
```

## Usage

### Server Mode

Start the server (default port 5201):
```bash
./iperf3-go
```

Start server on specific port:
```bash
./iperf3-go -p 8080
```

Start server with verbose output:
```bash
./iperf3-go -v
```

Handle only one client connection then exit:
```bash
./iperf3-go -1
```

Bind to specific interface:
```bash
./iperf3-go -B 192.168.1.100
```

### Client Mode

Basic TCP test (10 seconds):
```bash
./iperf3-go -c <server-ip>
```

TCP test for 30 seconds:
```bash
./iperf3-go -c <server-ip> -t 30
```

TCP test with multiple parallel streams:
```bash
./iperf3-go -c <server-ip> -P 4
```

TCP test with specific window size:
```bash
./iperf3-go -c <server-ip> -w 65536
```

TCP reverse test (server sends data to client):
```bash
./iperf3-go -c <server-ip> -R
```

JSON output format:
```bash
./iperf3-go -c <server-ip> -J
```

Custom bandwidth limit:
```bash
./iperf3-go -c <server-ip> -b 100000000
```

### UDP Mode

UDP server:
```bash
./iperf3-go -u -v
```

UDP client test:
```bash
./iperf3-go -c <server-ip> -u
```

UDP test with bandwidth limit (1 Mbps):
```bash
./iperf3-go -c <server-ip> -u -b 1000000
```

UDP test with custom packet size:
```bash
./iperf3-go -c <server-ip> -u -l 1470
```

### SCTP Mode

**Note**: SCTP requires Linux kernel support and is not available on Windows or macOS.

SCTP server (Linux only):
```bash
./iperf3-go -sctp -v
```

SCTP client test (Linux only):
```bash
./iperf3-go -c <server-ip> -sctp
```

SCTP test with multiple streams:
```bash
./iperf3-go -c <server-ip> -sctp -P 4
```

### Testing with Standard iperf3

You can also test interoperability with standard iperf3:

```bash
# Test iperf3-go server with standard iperf3 client
./iperf3-go -v &
iperf3 -c localhost -t 10

# Test iperf3-go client with standard iperf3 server
iperf3 -s &
./iperf3-go -c localhost -t 10
```

## Command Line Options

### Common Options
- `-p <port>`: Server port to listen on/connect to (default: 5201)
- `-v`: Verbose output
- `--version`: Show version information and quit

### Client Mode Options
- `-c <host>`: Run in client mode, connecting to `<host>`
- `-t <time>`: Time in seconds to transmit for (default: 10)
- `-P <streams>`: Number of parallel client streams to run (default: 1)
- `-R`: Run in reverse mode (server sends, client receives)
- `-J`: Output in JSON format
- `-w <window>`: Window size / socket buffer size
- `-l <length>`: Length of buffer to read or write (default: 128KB)
- `-b <bandwidth>`: Target bandwidth in bits/sec (0 for unlimited)
- `-u`: Use UDP rather than TCP
- `-sctp`: Use SCTP rather than TCP (Linux only)

### Server Mode Options
- `-B <host>`: Bind to a specific interface
- `-D`: Run the server as a daemon
- `-1`: Handle one client connection then exit

## Protocol Compatibility

This implementation follows the iperf3 protocol specification and is compatible with:
- iperf3 version 3.x clients
- Standard TCP performance tests
- JSON output format
- Real-time interval reporting

## Example Output

When running a test, you'll see output similar to the standard iperf3:

```
Server listening on 5201
New connection from 192.168.1.100:54321
Test config: {Protocol:tcp Time:10 Parallel:1 ...}
```

The client will receive standard iperf3 formatted results:

```
Connecting to host 192.168.1.1, port 5201
[  5] local 192.168.1.100 port 54321 connected to 192.168.1.1 port 5201
[ ID] Interval           Transfer     Bitrate
[  5]   0.00-1.00   sec  1.12 GBytes  9.65 Gbits/sec
[  5]   1.00-2.00   sec  1.15 GBytes  9.89 Gbits/sec
...
```

## Current Limitations

- UDP testing is not yet implemented
- Some advanced features may not be fully supported
- CPU utilization reporting is basic
- Reverse mode testing needs refinement

## Architecture

The server is structured with the following components:

- `main.go`: Entry point and command-line parsing
- `internal/server/`: Server implementation and session management
- `internal/protocol/`: iperf3 protocol message handling and data structures

## Testing

To test the server:

1. Build the server: `go build -o iperf3-go main.go`
2. Start the server: `./iperf3-go -v`
3. In another terminal, run: `iperf3 -c localhost -t 5`

## Troubleshooting

### "go: command not found"
- Install Go from https://golang.org/dl/
- Make sure Go is in your PATH

### "connection refused"
- Check if the server is running
- Verify the port number (default 5201)
- Check firewall settings

### "protocol error"
- Ensure you're using iperf3 client (not iperf2)
- Check server logs with `-v` flag

## Contributing

This is a basic implementation that covers the core iperf3 server functionality. Contributions are welcome to add:
- UDP support
- Additional protocol features
- Performance optimizations
- Better error handling
- Reverse mode improvements

## License

This project is licensed under the terms of the MIT license.
