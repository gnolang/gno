#!/usr/bin/env python3
import socket
import json
import time
import threading

def read_messages(sock):
    """Background thread to read messages from DAP server"""
    while True:
        try:
            # Read headers
            headers = b""
            while True:
                data = sock.recv(1)
                if not data:
                    return
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
            message = json.loads(body)
            
            # Print the message
            if message.get("type") == "event":
                print(f"\n🔔 EVENT: {message.get('event')}")
                if message.get('body'):
                    print(f"   Body: {json.dumps(message['body'], indent=2)}")
            elif message.get("type") == "response":
                print(f"\n✓ Response: {message.get('command')} (success={message.get('success')})")
                
        except Exception as e:
            print(f"Reader error: {e}")
            return

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
    print(f"→ Sent: {command}")

def main():
    # Connect to DAP server
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect(("localhost", 2345))
    print("Connected to DAP server\n")
    
    # Start reader thread
    reader = threading.Thread(target=read_messages, args=(sock,))
    reader.daemon = True
    reader.start()
    
    # Send initialize
    send_dap_request(sock, "initialize", {
        "clientID": "test",
        "clientName": "Test Client",
        "adapterID": "gno",
        "linesStartAt1": True,
        "columnsStartAt1": True
    })
    time.sleep(0.5)
    
    # Send launch
    send_dap_request(sock, "launch", {"program": "sample.gno"})
    time.sleep(0.5)
    
    # Set breakpoint
    send_dap_request(sock, "setBreakpoints", {
        "source": {"path": "sample.gno"},
        "breakpoints": [{"line": 37}]
    })
    time.sleep(0.5)
    
    # Configuration done
    send_dap_request(sock, "configurationDone")
    time.sleep(0.5)
    
    # Continue
    send_dap_request(sock, "continue", {"threadId": 1})
    
    # Wait for events
    print("\nWaiting for stopped event...")
    time.sleep(5)
    
    # Disconnect
    send_dap_request(sock, "disconnect")
    time.sleep(0.5)
    
    sock.close()
    print("\nDisconnected")

if __name__ == "__main__":
    main()