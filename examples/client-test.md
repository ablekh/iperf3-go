# iperf3-go Client Testing Examples

This document provides examples of how to test the iperf3-go server with various iperf3 client configurations.

## Basic Tests

### Simple TCP Test (10 seconds)
```bash
iperf3 -c localhost
```

### TCP Test with Custom Duration
```bash
iperf3 -c localhost -t 30
```

### TCP Test with JSON Output
```bash
iperf3 -c localhost -J
```

## Advanced Tests

### Multiple Parallel Streams
```bash
iperf3 -c localhost -P 4
```

### Custom Window Size
```bash
iperf3 -c localhost -w 64K
```

### Reverse Mode (Server sends to client)
```bash
iperf3 -c localhost -R
```

### Bidirectional Test
```bash
iperf3 -c localhost --bidir
```

### Custom Port
```bash
iperf3 -c localhost -p 8080
```

## Expected Output

When running a basic test, you should see output similar to:

```
Connecting to host localhost, port 5201
[  5] local 127.0.0.1 port 54321 connected to 127.0.0.1 port 5201
[ ID] Interval           Transfer     Bitrate
[  5]   0.00-1.00   sec  1.12 GBytes  9.65 Gbits/sec
[  5]   1.00-2.00   sec  1.15 GBytes  9.89 Gbits/sec
[  5]   2.00-3.00   sec  1.13 GBytes  9.71 Gbits/sec
[  5]   3.00-4.00   sec  1.14 GBytes  9.82 Gbits/sec
[  5]   4.00-5.00   sec  1.12 GBytes  9.64 Gbits/sec
[  5]   5.00-6.00   sec  1.15 GBytes  9.88 Gbits/sec
[  5]   6.00-7.00   sec  1.13 GBytes  9.73 Gbits/sec
[  5]   7.00-8.00   sec  1.14 GBytes  9.81 Gbits/sec
[  5]   8.00-9.00   sec  1.12 GBytes  9.65 Gbits/sec
[  5]   9.00-10.00  sec  1.15 GBytes  9.87 Gbits/sec
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate
[  5]   0.00-10.00  sec  11.4 GBytes  9.78 Gbits/sec                  sender
[  5]   0.00-10.00  sec  11.4 GBytes  9.78 Gbits/sec                  receiver

iperf Done.
```

## Troubleshooting

### Connection Refused
- Make sure the iperf3-go server is running
- Check that you're using the correct port (default 5201)
- Verify firewall settings

### Protocol Errors
- Ensure you're using iperf3 client (not iperf2)
- Check server logs with `-v` flag for detailed output

### Performance Issues
- Try different window sizes with `-w`
- Test with multiple streams using `-P`
- Check network configuration and hardware limitations