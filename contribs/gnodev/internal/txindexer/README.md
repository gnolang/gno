# Transaction Indexer Package

The `txindexer` package provides functionality to manage and control a
tx-indexer process for gnodev development. It handles the lifecycle of the
tx-indexer process, including starting, stopping, reloading the service. 
Reloading comprises stopping the current process, removing the database, and
starting it again. It's important to note that since the tx-indexer is in a
separate repo we use exec.Cmd to manage the process. This functionality is
provided by a `Service` for which an example of its usage is provided below. 

## Usage

### Basic Service Example

```go
// Configure the tx-indexer
cfg := txindexer.Config{
    Enabled:       true,
    DBPath:        "/path/to/db",
}

// Create a new service
svc, err := txindexer.NewService(logger, cfg)
if err != nil {
    log.Fatal(err)
}

// Start the tx-indexer
ctx := context.Background()
if err := svc.Start(ctx); err != nil {
    log.Fatal(err)
}

// Reload the tx-indexer (stops, removes DB, and starts again)
if err := svc.Reload(ctx); err != nil {
    log.Fatal(err)
}
```

## Logging

The tx-indexer process's stdout and stderr output is automatically piped to the
standard logger used across the gnodev application.

## Dependencies

- Requires the `tx-indexer` binary to be available in the system PATH
- Uses the standard library's `exec` package for process management. Non-Unix 
users will not be able to leverage the signal calls used for managing the
tx-indexer process.
