---
id: network-config
---

# Network configurations

| Network     | RPC Endpoint                      | Chain ID      | 
|-------------|-----------------------------------|---------------|
| Portal Loop | https://rpc.gno.land:443          | `portal-loop` |
| Testnet 4   | upcoming                          | upcoming      |
| Testnet 3   | https://rpc.test3.gno.land:443    | `test3`       |
| Staging     | http://rpc.staging.gno.land:36657 | `staging`     |

### WebSocket endpoints
All networks follow the same pattern for websocket connections: 

```shell
wss://<rpc-endpoint:port>/websocket
```