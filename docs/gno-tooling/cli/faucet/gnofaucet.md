---
id: gno-tooling-gnofaucet
---

# gnofaucet

`gnofaucet` is a server for distributing GNOT, the gas currency of Gnoland, to specific addresses in a local chain.
Interact with the `gnofaucet` from an address with an empty balance in your locally built testnet to fuel it with GNOT
to pay for transactions.

## Run `gnofaucet` Commands

Enable the faucet using the following command.

```bash
gnofaucet serve
```

#### **Options**

| Name                      | Type    | Description                                                                          |
|---------------------------|---------|--------------------------------------------------------------------------------------|
| `chain-id`                | String  | The id of the chain (required).                                                      |
| `gas-wanted`              | Int64   | The maximum amount of gas to use for the transaction (default: `50000`)              |
| `gas-fee`                 | String  | The gas fee to pay for the transaction.                                              |
| `memo`                    | String  | Any descriptive text (default: `""`)                                                 |
| `test-to`                 | String  | Test address (optional)                                                              |
| `send`                    | String  | Coins to send (default: `"1000000ugnot"`).                                           |
| `captcha-secret`          | String  | The secret key for the recaptcha. If empty, the captcha is disabled (default: `""`). |
| `is-behind-proxy`         | Boolean | Uses X-Forwarded-For IP for throttling (default: `false`).                           |
| `insecure-password-stdin` | Boolean | INSECURE! Takes password from stdin (default: `false`).                              |

## Example

### Step 1. Create an account named `test1` with the test seed phrase below.

```bash
gnokey add test1 --recover
```

> **Test Seed Phrase:** source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate
> oppose farm nothing bullet exhibit title speed wink action roast
> **Test Private key:** ea97b9fddb7e6bf6867090a7a819657047949fbb9466d617f940538efd888605
### **Step 2. Run `gnofaucet`**

```bash
gnofaucet serve test1 --chain-id dev --send 500000000ugnot
```

### **Step 3. Receive GNOTs from the faucet**

To receive funds through the `gnoweb` form GUI, you can request them on:
`http://localhost:8888/faucet` (given `http://localhost:8888/` is the location where `gnoweb` is serving pages).

Alternatively, you can request funds from the faucet by directly invoking a CURL command:

```bash
curl --location --request POST 'http://localhost:5050' \
--header 'Content-Type: application/x-www-form-urlencoded' \
--data-urlencode 'toaddr={address to receive}'
```
