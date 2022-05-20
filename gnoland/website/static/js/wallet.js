var remote = "http://gno.land:36657",
    chainId = "testchain",
    walletFn = {keplr: {}};

walletFn.createVmCallMsg = function (sender, pkgPath, funcName, args) {
	return  {
		type: "/vm.m_call",
		value: {
			caller: sender,
			send: "",
			pkg_path: pkgPath,
			func: funcName,
			args: args,
		}
	};
};

walletFn.createSignDoc = function(account, msg, chainId, gas) {
	return {
	  msgs: [msg],
	  fee: { amount: [{
		amount: "1",
		denom: "gnot"
	  }], gas: gas },
	  chain_id: chainId,
	  memo: "",
	  account_number: account.account_number,
	  sequence: account.sequence,
	};
};

walletFn.keplr.signAndBroadcast = function(sender, msg) {
	return window.keplr.experimentalSuggestChain(walletFn.getTestnetKeplrConfig())
	.then(function () {
		return window.keplr.enable(chainId);
	})
	.then(function () {
		return walletFn.getAccount(sender);
	})
	.then(function(account) {
		const signDoc = walletFn.createSignDoc(account, msg, chainId, "2000000");
		return window.keplr.signAmino(chainId, sender, signDoc, {
			// use app fee (1gnot fixed fee)
			preferNoSetFee: true, 
		});
	})
	.then(function (signature) {
		const tx = gnopb.makeProtoTx(signature.signed, signature.signature);
		return walletFn.broadcastTx(tx);
	});
};

walletFn.getTestnetKeplrConfig = function() {
	const addressPrefix = "g";
	const gnoToken = {
		coinDenom: "GNOT",
		coinMinimalDenom: "gnot",
		coinDecimals: 6,
		// coinGeckoId: ""
	};

	return {
		chainId: chainId,
		chainName: "GNO Testnet",
		rpc: 'http://gno.land:36657',
		rest: 'https://lcd.gno.tools',  // source: https://github.com/disperze/gno-api
		bech32Config: {
			bech32PrefixAccAddr: `${addressPrefix}`,
			bech32PrefixAccPub: `${addressPrefix}pub`,
			bech32PrefixValAddr: `${addressPrefix}valoper`,
			bech32PrefixValPub: `${addressPrefix}valoperpub`,
			bech32PrefixConsAddr: `${addressPrefix}valcons`,
			bech32PrefixConsPub: `${addressPrefix}valconspub`,
		},
		currencies: [gnoToken],
		feeCurrencies: [gnoToken],
		stakeCurrency: gnoToken,
		gasPriceStep: {
			low: 0.000000001, // min 1gnot for any tx
			average: 0.000000001,
			high: 0.000000001,
		},
		bip44: { coinType: 118 },
		// custom feature for GNO chains.
		features: ["gno"]
	};
};

walletFn.getAccount = function(address) {
	return walletFn.rpcCall("abci_query", {path: `auth/accounts/${address}`})
	.then(function (data) {
		const response = data.result.response.ResponseBase;
		if (response.Error) {
			throw new Error(response.Log);
		}

		const account = JSON.parse(atob(response.Data));
		if (!account) {
			throw new Error("Account not found");
		}

		return account.BaseAccount;
	});
};

walletFn.broadcastTx = function(tx) {
	return walletFn.rpcCall("broadcast_tx_commit", {tx: walletFn.tob64(tx)}).then(function (data) {
		if (data.error) {
			throw new Error(data.error.message);
		}
		
		const checkTx = data.result.check_tx.ResponseBase;
		if (checkTx.Error) {
			throw new Error(checkTx.Log);
		}

		const deliverTx = data.result.deliver_tx.ResponseBase;
		if (deliverTx.Error) {
			throw new Error(deliverTx.Log);
		}

		return {
			height: data.result.height,
			txhash: walletFn.base64ToHex(data.result.hash),
			gasWanted: deliverTx.GasWanted,
			gasUsed: deliverTx.GasUsed,
		};
	});
};

walletFn.tob64 = function(data) {
	return btoa(String.fromCharCode.apply(null, data));
};

walletFn.base64ToHex = function(data) {
	return atob(data)
		.split('')
		.map(function (c) {
		  return ('0' + c.charCodeAt(0).toString(16)).slice(-2);
		})
	   .join('')
	   .toUpperCase();
}

walletFn.rpcCall = function(method, params) {
	const payload  = {
		"jsonrpc": "2.0",
		"method": method,
		"params": params,
		"id": 1
	};

	return fetch(remote, {
		method: "POST",
		headers: {
			"Content-Type": "application/json"
		},
		body: JSON.stringify(payload)
	}).then(function (response) {
		return response.json();
	});
}
