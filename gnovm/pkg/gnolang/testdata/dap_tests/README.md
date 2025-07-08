# DAP (Debug Adapter Protocol) Test Suite

This directory contains test files and utilities for testing the DAP server implementation in Gno debugger.

## Files

### Test Programs
- `test_debug.gno` - Sample Gno program with various constructs for testing debugger features

### Python Test Clients
- `test_dap_client.py` - Basic DAP client for testing initialize/launch sequence
- `test_dap_breakpoint.py` - DAP client for testing breakpoint functionality
- `test_dap_simple.py` - Simplified DAP client with event monitoring

### Debug Command Files
- `debug_commands.txt` - Basic debugger commands for CLI testing
- `debug_commands2.txt` - Breakpoint-specific commands for CLI testing
- `debug_test.txt` - Additional test commands

## Usage

### Running DAP Server

From the debugger directory:

```bash
# Start DAP server on TCP port
../../../build/gno run --debug-addr localhost:2345 --dap sample.gno

# Or with test program
../../../build/gno run --debug-addr localhost:2345 --dap testdata/dap_tests/test_debug.gno
```

### Running Test Clients

```bash
# Basic connection test
python3 testdata/dap_tests/test_dap_client.py

# Breakpoint test
python3 testdata/dap_tests/test_dap_breakpoint.py

# Simple event monitoring
python3 testdata/dap_tests/test_dap_simple.py
```

### CLI Debugger Testing

```bash
# Run with command file
../../../build/gno run --debug sample.gno < testdata/dap_tests/debug_commands.txt
```

## Test Coverage

The test suite covers:
- DAP server initialization
- Client connection handling
- Breakpoint setting and hitting
- Stopped event generation
- Variable inspection
- Stack trace retrieval
- Continue/Step operations
- Disconnect handling