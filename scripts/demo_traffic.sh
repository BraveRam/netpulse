#!/usr/bin/env bash
# scripts/demo_traffic.sh - Generates network traffic to demonstrate netpulse radar features

set -e

echo "Starting synthetic network traffic generator for netpulse..."
echo "Press Ctrl+C to stop."

# Start a background TCP echo server
python3 -c "
import socket, time
s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(('127.0.0.1', 8999))
s.listen(5)
print('Demo server listening on 127.0.0.1:8999')
while True:
    conn, addr = s.accept()
    data = conn.recv(1024)
    conn.sendall(b'HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nOK')
    conn.close()
" &
SERVER_PID=$!

trap "kill $SERVER_PID 2>/dev/null || true" EXIT

while true; do
    # Trigger DNS queries and connections to various public endpoints
    curl -s --connect-timeout 2 http://127.0.0.1:8999 > /dev/null 2>&1 || true
    curl -s --connect-timeout 2 https://1.1.1.1 > /dev/null 2>&1 || true
    sleep 1
done
