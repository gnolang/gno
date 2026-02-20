#!/usr/bin/env python3
import socket
import json
import time

def send_dap_request(sock, command, args=None):
    request = {
        "seq": 1,
        "type": "request",
        "command": command
    }
    if args:
        request["arguments"] = args
    
    body = json.dumps(request)
    header = f"Content-Length: {len(body)}\r\n\r\n"
    message = header + body
    sock.sendall(message.encode())
    print(f"Sent: {command}")

def read_dap_message(sock):
    # Read headers
    headers = b""
    while True:
        data = sock.recv(1)
        if not data:
            return None
        headers += data
        if headers.endswith(b"\r\n\r\n"):
            break
    
    # Parse content length
    header_str = headers.decode()
    content_length = 0
    for line in header_str.split("\r\n"):
        if line.startswith("Content-Length:"):
            content_length = int(line.split(":")[1].strip())
            break
    
    # Read body
    body = b""
    while len(body) < content_length:
        chunk = sock.recv(content_length - len(body))
        if not chunk:
            break
        body += chunk
    
    return json.loads(body.decode())

# Connect
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(("localhost", 2345))
print("Connected to DAP server")

# Initialize
send_dap_request(sock, "initialize", {
    "clientID": "test",
    "adapterID": "gno"
})

# Read response
msg = read_dap_message(sock)
print(f"Response: {msg['type']} - {msg.get('command', msg.get('event', '?'))}")

# Close
sock.close()
print("Done")