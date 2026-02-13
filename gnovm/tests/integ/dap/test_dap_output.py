#!/usr/bin/env python3
import socket
import json
import time

class DAPClient:
    def __init__(self, host, port):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((host, port))
        self.seq = 0
        self.output_messages = []
        
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
                event_type = msg.get('event')
                print(f"← Event: {event_type}")
                if event_type == "output":
                    output = msg.get('body', {}).get('output', '')
                    category = msg.get('body', {}).get('category', 'unknown')
                    self.output_messages.append((category, output))
                    print(f"  Output ({category}): {repr(output)}")
        return None
        
    def close(self):
        self.sock.close()

def test_output_events():
    client = DAPClient("localhost", 2345)
    print("Connected to DAP server\n")
    
    try:
        # Initialize
        seq = client.send_request("initialize", {
            "clientID": "test-output",
            "clientName": "Output Test Client",
            "adapterID": "gno",
            "linesStartAt1": True,
            "columnsStartAt1": True
        })
        resp = client.wait_for_response(seq)
        print(f"← Initialize: {resp['success']}")
        
        # Launch
        seq = client.send_request("launch", {
            "program": "../../tests/integ/debugger/test_output.gno"
        })
        resp = client.wait_for_response(seq)
        print(f"← Launch: {resp['success']}")
        
        # Don't set breakpoints - let program run to completion
        
        # Configuration done
        seq = client.send_request("configurationDone")
        resp = client.wait_for_response(seq)
        
        # Wait for output events
        print("\nWaiting for output events...")
        time.sleep(2)
        
        # Give a bit more time for all output events
        time.sleep(1)
        
        # Read any remaining messages
        try:
            while True:
                msg = client.read_message()
                if msg.get("type") == "event" and msg.get("event") == "output":
                    output = msg.get('body', {}).get('output', '')
                    category = msg.get('body', {}).get('category', 'unknown')
                    client.output_messages.append((category, output))
                    print(f"  Output ({category}): {repr(output)}")
        except:
            pass
        
        # Print collected output
        print(f"\nCollected {len(client.output_messages)} output messages:")
        for category, output in client.output_messages:
            print(f"  [{category}] {repr(output)}")
        
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
    test_output_events()