#!/usr/bin/env python3
import socket
import json

def send_dap_request(sock, command, args=None):
    seq_num = 1
    request = {
        "seq": seq_num,
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

def read_dap_response(sock):
    # Read headers
    headers = b""
    while True:
        data = sock.recv(1)
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
    body = sock.recv(content_length)
    response = json.loads(body)
    print(f"Received: {json.dumps(response, indent=2)}")
    return response

def main():
    # Connect to DAP server
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(("localhost", 2345))
    print("Connected to DAP server")
    
    # Send initialize request
    init_args = {
        "clientID": "test-client",
        "clientName": "Test DAP Client",
        "adapterID": "gno",
        "linesStartAt1": True,
        "columnsStartAt1": True
    }
    send_dap_request(sock, "initialize", init_args)
    read_dap_response(sock)
    
    # Send launch request
    launch_args = {
        "program": "test_debug.gno"
    }
    send_dap_request(sock, "launch", launch_args)
    read_dap_response(sock)
    
    # Close connection
    sock.close()
    print("Disconnected from DAP server")

if __name__ == "__main__":
    main()