# iperf3-go

A Go implementation of an iperf3 server that is compatible with the standard Linux iperf3 client.

## Features

- TCP performance testing
- Compatible with standard iperf3 clients
- JSON output format matching iperf3
- Configurable server options
- Real-time interval reporting
- Multiple client support

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

### Client Testing

Use the standard iperf3 client to test against this server:

```bash
# Basic TCP test (10 seconds)
iperf3 -c <server-ip>

# TCP test for 30 seconds
iperf3 -c <server-ip> -t 30

# TCP test with multiple parallel streams
iperf3 -c <server-ip> -P 4

# TCP test with specific window size
iperf3 -c <server-ip> -w 64K

# TCP reverse test (server sends data to client)
iperf3 -c <server-ip> -R

# JSON output format
iperf3 -c <server-ip> -J
```

## Command Line Options

- `-p <port>`: Server port to listen on (default: 5201)
- `-B <host>`: Bind to a specific interface
- `-v`: Verbose output
- `-D`: Run the server as a daemon
- `-1`: Handle one client connection then exit
- `--version`: Show version information and quit

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

This project is provided as-is for educational and testing purposes.