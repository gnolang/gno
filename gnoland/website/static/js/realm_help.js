function main() {
	// init
	var myAddr = getMyAddress()
	u("#my_address").first().value = myAddr;
	setMyAddress(myAddr);
	// main renders
	u("div.func_spec").each(function(x) {
		updateCommand(u(x));
	});
	// main hooks
	u("div.func_spec input").on("input", function(e) {
		var x = u(e.currentTarget).closest("div.func_spec");
		updateCommand(x);
	});
	// special case: when address changes.
	u("#my_address").on("input", function(e) {
		var value = u("#my_address").first().value;
		setMyAddress(value)
		u("div.func_spec").each(function(node, i){
			updateCommand(u(node));
		});
	});

	u(".keplr_exec").on("click", function(e) {
		var el = u(e.currentTarget);
		var sender = el.data("sender");
		var msg = el.data("msg");
		sendTx(sender, JSON.parse(msg));
	});
};

function setMyAddress(addr) {
	localStorage.setItem("my_address", addr);
}

function getMyAddress() {
	var myAddr = localStorage.getItem("my_address");
	if (!myAddr){
		return "ADDRESS";
	}
	return myAddr;
}

// x: the u("div.func_spec") element.
function updateCommand(x) {
	var realmPath = u("#data").data("realm-path");
	var funcName = x.data("func-name");
	var ins = x.find("table>tbody>tr.func_params input");
	var vals = [];
	ins.each(function(input) {
		vals.push(input.value);
	});
	var myAddr = getMyAddress();
	var shell = x.find(".shell_command");
	shell.empty();

	// command Z: all in one.
	shell.append(u("<span>").text("### INSECURE BUT QUICK ###")).append(u("<br>"));
	var args = ["gnokey", "maketx", "call", myAddr,
		"--pkgpath", shq(realmPath), "--func", shq(funcName),
		"--gas-fee", "1gnot", "--gas-wanted", "2000000",
		"--send", shq(""),
		"--broadcast", "true", "--chainid", "testchain"];
	vals.forEach(function(arg) {
		args.push("--args");
		args.push(shq(arg));
	});
	args.push("--remote", "gno.land:36657");
	var command = args.join(" ");
	shell.append(u("<span>").text(command)).append(u("<br>")).append(u("<br>"));

	// or...
	shell.append(u("<span>").text("### FULL SECURITY WITH AIRGAP ###")).append(u("<br>"));

	// command 0: query account info.
	var args = ["gnokey", "query", "auth/accounts/" + myAddr, "--remote", "gno.land:36657"];
	var command = args.join(" ");
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// command 1: construct tx.
	var args = ["gnokey", "maketx", "call", myAddr,
		"--pkgpath", shq(realmPath), "--func", shq(funcName),
		"--gas-fee", "1gnot", "--gas-wanted", "2000000",
		"--send", shq("")];
	vals.forEach(function(arg) {
		args.push("--args");
		args.push(shq(arg));
	});
	var command = args.join(" ");
	command = command+" > unsigned.tx";
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// command 2: sign tx.
	var args = ["gnokey", "sign", myAddr,
		"--txpath", "unsigned.tx", "--chainid", "testchain",
		"--number", "ACCOUNTNUMBER",
		"--sequence", "SEQUENCENUMBER"];
	var command = args.join(" ");
	command = command+" > signed.tx";
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// command 3: broadcast tx.
	var args = ["gnokey", "broadcast", "signed.tx", "--remote", "gno.land:36657"];
	var command = args.join(" ");
	command = command;
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// set keplr params
	var keplrExec = x.find(".keplr_exec");
	const msg = createMsg(myAddr, realmPath, funcName, vals);
	keplrExec.data({msg: JSON.stringify(msg), sender: myAddr});
}

function createMsg(sender, pkgPath, funcName, args) {
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
}

function createSignDoc(account, msg, chainId, gas) {
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
  }

function sendTx(sender, msg) {
	// validate params

	// get account info.
	var account = {account_number: 0, sequence: 0};
	const signDoc = createSignDoc(account, msg, "testchain", "2000000");
	const chainId = "testchain";
	window.keplr.enable(chainId).then(function () {
		return keplr.signAmino(chainId, sender, signDoc, {
			preferNoSetFee: true, // 1gnot fixed fee 
		});
	}).then(function (signature) {
		const tx = makeStdTx(signature.signed, signature.signature);
		return broadcastTx(tx);
	}).then(function (result) {
		console.log(result.txHash);
	}).catch(function (err) {
		console.log(err);
	});
}

function makeStdTx(content, signature) {
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
}

function broadcastTx(tx) {
	return fetch("/txs", {
		method: "POST",
		headers: {
			"Content-Type": "application/json"
		},
		body: JSON.stringify(tx)
	}).then(function (response) {
		return response.json();
	});
}

// Jae: why isn't this a library somewhere?
function shq(s) {
	var s2 = String(s).replace(/\t/g, '\\t');
	var s2 = String(s2).replace(/\n/g, '\\n');
	var s2 = String(s2).replace(/([$'"`\\!])/g, '\\$1'); 
	return '"'+s2+'"';
};
