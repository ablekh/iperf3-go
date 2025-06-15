#!/bin/bash

# Test script for iperf3-go server

echo "Building iperf3-go server..."
go build -o iperf3-go main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Starting iperf3-go server in background..."
./iperf3-go -v &
SERVER_PID=$!

# Give server time to start
sleep 2

echo "Testing with iperf3 client..."
echo "Running: iperf3 -c localhost -t 5"

# Test if iperf3 client is available
if command -v iperf3 &> /dev/null; then
    iperf3 -c localhost -t 5
    TEST_RESULT=$?
else
    echo "iperf3 client not found. Please install iperf3 to test:"
    echo "  Ubuntu/Debian: sudo apt-get install iperf3"
    echo "  CentOS/RHEL: sudo yum install iperf3"
    echo "  macOS: brew install iperf3"
    echo "  Windows: Download from https://iperf.fr/iperf-download.php"
    TEST_RESULT=1
fi

echo "Stopping server..."
kill $SERVER_PID

if [ $TEST_RESULT -eq 0 ]; then
    echo "Test completed successfully!"
else
    echo "Test failed or iperf3 client not available"
fi

exit $TEST_RESULT