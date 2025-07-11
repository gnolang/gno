#!/usr/bin/env python3
import socket
import json
import time
import sys

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
        print(f"→ {command}")
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
            
        return json.loads(body)
    
    def wait_for_response(self, req_seq, timeout=5):
        """Wait for a specific response"""
        start_time = time.time()
        while time.time() - start_time < timeout:
            msg = self.read_message()
            if msg.get("type") == "response" and msg.get("request_seq") == req_seq:
                return msg
            elif msg.get("type") == "event":
                print(f"← Event: {msg.get('event')}")
                if msg.get('event') == 'output':
                    output = msg.get('body', {}).get('output', '')
                    print(f"  Output: {repr(output)}")
        return None
        
    def close(self):
        self.sock.close()

def test_attach_mode():
    """
    Test attach mode functionality.
    
    Before running this test, start the DAP server in attach mode:
    gno run --debug-addr localhost:2345 --dap --attach gnovm/tests/integ/debugger/test_output.gno
    """
    client = DAPClient("localhost", 2345)
    print("Connected to DAP server in attach mode\n")
    
    try:
        # Initialize
        seq = client.send_request("initialize", {
            "clientID": "test-attach",
            "clientName": "Attach Test Client",
            "adapterID": "gno",
            "linesStartAt1": True,
            "columnsStartAt1": True
        })
        resp = client.wait_for_response(seq)
        print(f"← Initialize: {resp['success']}")
        
        # Attach (instead of launch)
        seq = client.send_request("attach", {
            "port": 2345,
            "host": "localhost"
        })
        resp = client.wait_for_response(seq)
        print(f"← Attach: {resp['success']}")
        
        # Set breakpoints before starting
        seq = client.send_request("setBreakpoints", {
            "source": {"path": "test_output.gno"},
            "breakpoints": [{"line": 5}]  # println("Testing output redirection")
        })
        resp = client.wait_for_response(seq)
        print(f"← SetBreakpoints: {resp['success']}")
        
        # Configuration done - program should start executing now
        seq = client.send_request("configurationDone")
        resp = client.wait_for_response(seq)
        print(f"← ConfigurationDone: {resp['success']}")
        
        # Wait for stopped event at breakpoint
        print("\nWaiting for program to hit breakpoint...")
        time.sleep(1)
        
        # Continue execution
        seq = client.send_request("continue", {"threadId": 1})
        resp = client.wait_for_response(seq)
        
        # Wait for program output
        print("\nWaiting for program output...")
        time.sleep(2)
        
        # Disconnect
        seq = client.send_request("disconnect")
        client.wait_for_response(seq, timeout=1)
        
    except Exception as e:
        print(f"Error: {e}")
        import traceback
        traceback.print_exc()
    finally:
        client.close()
        print("\nDisconnected")

if __name__ == "__main__":
    print("Testing DAP Attach Mode")
    print("=" * 50)
    print("Make sure to start the DAP server with:")
    print("gno run --debug-addr localhost:2345 --dap --attach gnovm/tests/integ/debugger/test_output.gno")
    print("=" * 50)
    print()
    
    # Give user time to read instructions
    time.sleep(2)
    
    test_attach_mode()