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
        return None
        
    def close(self):
        self.sock.close()

def test_variable_expansion():
    client = DAPClient("localhost", 2345)
    print("Connected to DAP server\n")
    
    try:
        # Initialize
        seq = client.send_request("initialize", {
            "clientID": "test-variables",
            "clientName": "Variable Test Client",
            "adapterID": "gno",
            "linesStartAt1": True,
            "columnsStartAt1": True
        })
        resp = client.wait_for_response(seq)
        print(f"← Initialize: {resp['success']}")
        
        # Launch
        seq = client.send_request("launch", {
            "program": "../../tests/integ/debugger/test_struct.gno"
        })
        resp = client.wait_for_response(seq)
        print(f"← Launch: {resp['success']}")
        
        # Set breakpoint at line 15 (after creating struct)
        seq = client.send_request("setBreakpoints", {
            "source": {"path": "test_struct.gno"},
            "breakpoints": [{"line": 15}]  # println(p.Name)
        })
        resp = client.wait_for_response(seq)
        print(f"← SetBreakpoints: {resp['success']}")
        
        # Configuration done
        seq = client.send_request("configurationDone")
        resp = client.wait_for_response(seq)
        
        # Continue to breakpoint
        seq = client.send_request("continue", {"threadId": 1})
        resp = client.wait_for_response(seq)
        
        # Wait a bit for stopped event
        time.sleep(0.5)
        
        # Get threads
        seq = client.send_request("threads")
        resp = client.wait_for_response(seq)
        
        # Get stack trace
        seq = client.send_request("stackTrace", {"threadId": 1})
        resp = client.wait_for_response(seq)
        
        if resp and resp.get('body', {}).get('stackFrames'):
            frame_id = resp['body']['stackFrames'][0]['id']
            print(f"\nFrame ID: {frame_id}")
            
            # Get scopes
            seq = client.send_request("scopes", {"frameId": frame_id})
            resp = client.wait_for_response(seq)
            
            if resp and resp.get('body', {}).get('scopes'):
                # Get local variables
                for scope in resp['body']['scopes']:
                    if scope['name'] == 'Locals':
                        print(f"\nGetting variables for scope: {scope['name']}")
                        seq = client.send_request("variables", {
                            "variablesReference": scope['variablesReference']
                        })
                        resp = client.wait_for_response(seq)
                        
                        if resp and resp.get('body', {}).get('variables'):
                            print("\nLocal Variables:")
                            for var in resp['body']['variables']:
                                ref = var.get('variablesReference', 0)
                                print(f"  {var['name']} = {var['value']} ({var.get('type', 'unknown')}) [ref={ref}]")
                                
                                # If variable has children, expand it
                                if ref > 0:
                                    print(f"    Expanding {var['name']}...")
                                    seq = client.send_request("variables", {
                                        "variablesReference": var['variablesReference']
                                    })
                                    child_resp = client.wait_for_response(seq)
                                    
                                    if child_resp and child_resp.get('body', {}).get('variables'):
                                        for child in child_resp['body']['variables']:
                                            print(f"      {child['name']} = {child['value']} ({child.get('type', 'unknown')})")
        
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
    test_variable_expansion()