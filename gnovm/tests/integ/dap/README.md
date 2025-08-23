# DAP (Debug Adapter Protocol) Test Suite

This directory contains test files and utilities for testing the DAP server implementation in Gno debugger.

## Files

### Test Programs
- `test_debug.gno` - Sample Gno program with various constructs for testing debugger features
- `test_output.gno` - Test program for output redirection functionality
- `test_struct.gno` - Test program for struct variable expansion

### Python Test Clients
- `test_dap_client.py` - Basic DAP client for testing initialize/launch sequence
- `test_dap_breakpoint.py` - DAP client for testing breakpoint functionality
- `test_dap_simple.py` - Simplified DAP client with event monitoring
- `test_dap_variables.py` - Test for variable expansion functionality
- `test_dap_output.py` - Test for output event redirection
- `test_output_monitor.py` - Monitor for DAP output events
- `test_dap_attach.py` - Test for attach mode functionality

### Debug Command Files
- `debug_commands.txt` - Basic debugger commands for CLI testing
- `debug_commands2.txt` - Breakpoint-specific commands for CLI testing
- `debug_test.txt` - Additional test commands

## Usage

### Running DAP Server

From the project root directory:

```bash
# Start DAP server on TCP port (launch mode)
gno run --debug-addr localhost:2345 --dap gnovm/tests/integ/debugger/sample.gno

# Or with test program
gno run --debug-addr localhost:2345 --dap gnovm/tests/integ/dap/test_debug.gno

# Start DAP server in attach mode (waits for client to connect before executing)
gno run --debug-addr localhost:2345 --dap --attach gnovm/tests/integ/debugger/test_output.gno
```

### Running Test Clients

From the project root directory:

```bash
# Basic connection test
python3 gnovm/tests/integ/dap/test_dap_client.py

# Breakpoint test
python3 gnovm/tests/integ/dap/test_dap_breakpoint.py

# Simple event monitoring
python3 gnovm/tests/integ/dap/test_dap_simple.py

# Variable expansion test
python3 gnovm/tests/integ/dap/test_dap_variables.py

# Output redirection test
python3 gnovm/tests/integ/dap/test_dap_output.py

# Attach mode test (requires server started with --attach flag)
python3 gnovm/tests/integ/dap/test_dap_attach.py
```

### CLI Debugger Testing

From the project root directory:

```bash
# Run with command file
gno run --debug gnovm/tests/integ/debugger/sample.gno < gnovm/tests/integ/dap/debug_commands.txt
```

## Test Coverage

The test suite covers:
- DAP server initialization
- Client connection handling
- Launch and Attach modes
- Breakpoint setting and hitting
- Stopped event generation
- Variable inspection with complex type expansion
- Stack trace retrieval
- Continue/Step operations
- Output event redirection
- Program termination handling
- Disconnect handling
