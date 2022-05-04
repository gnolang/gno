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
	return window.keplr.experimentalSuggestChain(walletFn.geTestnetKeplrConfig())
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
		const tx = walletFn.makeStdTx(signature.signed, signature.signature);
		return walletFn.broadcastTx(tx);
	});
};

walletFn.geTestnetKeplrConfig = function() {
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
		rest: 'https://lcd.gno.tools',
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
		bip44: { coinType: 118 },
		// custom feature for GNO chains.
		features: ["gno"]
	};
};

walletFn.makeStdTx = function(content, signature) {
	const feeAmount = content.fee.amount;

	return {
		msg: content.msgs.map(function (msg) {
			return {
				"@type": msg.type,
				...msg.value
			};
		}),
		fee: {
			gas_wanted: content.fee.gas,
			gas_fee: feeAmount.length ? `${content.fee.amount[0].amount}${content.fee.amount[0].denom}`: "",
		},
		signatures: [{
			pub_key: {
				"@type": signature.pub_key.type,
				value: signature.pub_key.value,
			},
			signature: signature.signature,
		}],
		memo: content.memo,
	};
};

walletFn.getAccount = function(address) {
	return walletFn.abciQuery(`auth/accounts/${address}`)
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

walletFn.abciQuery = function(path) {
	return fetch(`${remote}/abci_query?path="${path}"`).then(function(response) {
		return response.json();
	});
};

walletFn.broadcastTx = function(tx) {
	const payload = {
		tx: tx,
	};
	return fetch("/txs", {
		method: "POST",
		headers: {
			"Content-Type": "application/json"
		},
		body: JSON.stringify(payload)
	}).then(function (response) {
		return response.json();
	});
};
