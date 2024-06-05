---
id: connect-wallet-dapp
---

# How to connect a wallet to a dApp

As a dapp developer, you must integrate a web3 wallet with your application to enable users to interact with your
application. Upon integration, you may retrieve account information of the connected user or request to sign & send
transactions from the user's account.

:::warning Wallets on gno.land

Here is a list of available wallets for Gnoland.
Note that none of these wallets are official or exclusive, so please
use them at your own diligence:

- [Adena Wallet](https://adena.app/)

:::

## Adena Wallet

[Adena](https://adena.app/) is a web extension wallet that supports the Gnoland blockchain. Below is the basic Adena
APIs that you can use for your application. For more detailed information, check out
Adena's [developer's docs](https://docs.adena.app/) to integrate Adena to your application.

### Adena Connect For React App

Check if Adena wallet exists.

```javascript
// checks the existence of the adena object in window

const existsWallet = () => {
    if (window.adena) {
        return true;
    }
    return false;
};

```

Register your website as a trusted domain.

```javascript
// calls the AddEstablish of the adena object

const addEstablish = (siteName) => {
    return window?.adena?.AddEstablish(siteName);
};

```

Retrieve information about the connected account.

```javascript
// calls the GetAccount function of the adena object

const getAccount = () => {
    return window.adena?.GetAccount();
};

```

Request approval of a transaction that transfers tokens.

```javascript
// Execute the DoContract function of the adena object to request transaction.

const sendToken = (fromAddress, toAddress, sendAmount) => {
    const message = {
        type: "/bank.MsgSend",
        value: {
            from_address: fromAddress,
            to_address: toAddress,
            amount: sendAmount
        }
    };

    return window.adena?.DoContract({
        messages: [message],
        gasFee: 1,
        gasWanted: 3000000
    });
};

```

Request approval of a transaction that calls a function from a realm.

```javascript
// Execute the DoContract function of the adena object to request transaction.

const doContractPackageFunction = (caller, func, pkgPath, argument) => {

    // Setup Transaction Message
    const message = {
        type: "/vm.m_call",
        value: {
            caller,
            func,
            send: "",
            pkg_path: pkgPath,
            args: argument.split(',')
        }
    };

    // Request Transaction
    return window.adena?.DoContract({
        messages: [message],
        gasFee: 1,
        gasWanted: 3000000
    });
};
```
