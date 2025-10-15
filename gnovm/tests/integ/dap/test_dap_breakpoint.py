#!/usr/bin/env python3
import socket
import json
import sys
import time

class DAPClient:
    def __init__(self, host, port):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((host, port))
        self.seq = 0
        
    def send_request(self, command, args=None):
        self.seq += 1
        request = {
            "seq": self.seq,
            "type": "request",
            "command": command
        }
        if args:
            request["arguments"] = args
        
        body = json.dumps(request)
        header = f"Content-Length: {len(body)}\r\n\r\n"
        message = header + body
        self.sock.sendall(message.encode())
        print(f"→ Sent {command} request")
        return self.seq
        
    def read_message(self):
        # Read headers
        headers = b""
        while True:
            data = self.sock.recv(1)
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
            chunk = self.sock.recv(content_length - len(body))
            if not chunk:
                break
            body += chunk
            
        message = json.loads(body)
        
        # Print the message based on type
        if message.get("type") == "response":
            print(f"← Response to {message.get('command')}: success={message.get('success')}")
            if message.get('body'):
                print(f"  Body: {json.dumps(message['body'], indent=2)}")
        elif message.get("type") == "event":
            print(f"← Event: {message.get('event')}")
            if message.get('body'):
                print(f"  Body: {json.dumps(message['body'], indent=2)}")
                
        return message
    
    def wait_for_event(self, event_type, timeout=5):
        """Wait for a specific event type"""
        start_time = time.time()
        while time.time() - start_time < timeout:
            msg = self.read_message()
            if msg.get("type") == "event" and msg.get("event") == event_type:
                return msg
        return None
        
    def close(self):
        self.sock.close()

def main():
    # Start DAP client
    client = DAPClient("localhost", 2345)
    print("Connected to DAP server\n")
    
    try:
        # 1. Initialize
        client.send_request("initialize", {
            "clientID": "test-client",
            "clientName": "Test DAP Client",
            "adapterID": "gno",
            "linesStartAt1": True,
            "columnsStartAt1": True
        })
        client.read_message()
        
        # 2. Launch
        client.send_request("launch", {
            "program": "test_debug.gno"
        })
        client.read_message()
        
        # Wait for initialized event
        initialized = client.wait_for_event("initialized", 2)
        if initialized:
            print("\n✓ Received initialized event")
        
        # 3. Set breakpoints
        print("\nSetting breakpoints...")
        client.send_request("setBreakpoints", {
            "source": {
                "name": "test_debug.gno",
                "path": "test_debug.gno"
            },
            "breakpoints": [
                {"line": 6},  # z := add(x, y)
                {"line": 15}, # slice = append(slice, "!")
            ]
        })
        client.read_message()
        
        # 4. Configuration done
        client.send_request("configurationDone")
        client.read_message()
        
        # 5. Continue execution
        print("\nContinuing execution...")
        client.send_request("continue", {"threadId": 1})
        client.read_message()
        
        # 6. Wait for stopped event at breakpoint
        print("\nWaiting for stopped event...")
        stopped = client.wait_for_event("stopped", 10)
        if stopped:
            print(f"\n✓ Hit breakpoint! Reason: {stopped['body']['reason']}")
            
            # 7. Get stack trace
            client.send_request("stackTrace", {"threadId": 1})
            stack_resp = client.read_message()
            
            # 8. Get variables
            if stack_resp.get('body', {}).get('stackFrames'):
                frame_id = stack_resp['body']['stackFrames'][0]['id']
                
                # Get scopes
                client.send_request("scopes", {"frameId": frame_id})
                scopes_resp = client.read_message()
                
                # Get local variables
                if scopes_resp.get('body', {}).get('scopes'):
                    for scope in scopes_resp['body']['scopes']:
                        if scope['name'] == 'Locals':
                            client.send_request("variables", {
                                "variablesReference": scope['variablesReference']
                            })
                            vars_resp = client.read_message()
                            
            # 9. Continue again
            print("\nContinuing to next breakpoint...")
            client.send_request("continue", {"threadId": 1})
            client.read_message()
            
            # Wait for next stopped event
            stopped2 = client.wait_for_event("stopped", 10)
            if stopped2:
                print(f"\n✓ Hit second breakpoint! Reason: {stopped2['body']['reason']}")
        else:
            print("\n✗ No stopped event received - breakpoint may not be working correctly")
            
        # 10. Disconnect
        print("\nDisconnecting...")
        client.send_request("disconnect")
        client.read_message()
        
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()
    finally:
        client.close()
        print("Disconnected from DAP server")

if __name__ == "__main__":
    main()