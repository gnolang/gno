function main() {
	u("div.func_spec").each(function(x) {
		updateCommand(u(x));
	});
	u("div.func_spec input").on("input", function(e) {
		var x = u(e.currentTarget).closest("div.func_spec");
		updateCommand(x);
	});
};

// x: the u("div.func_spec") element.
function updateCommand(x) {
	var realmPath = u("#data").data("realm-path");
	var funcName = x.data("func-name");
	var ins = x.find("table>tbody>tr.func_params input");
	var vals = [];
	ins.each(function(input) {
		vals.push(input.value);
	});
	var shell = x.find(".shell_command");
	shell.empty();

	// command 1: construct tx.
	var args = ["gnokey", "maketx", "call", "KEYNAME", 
		"--pkgpath", shq(realmPath), "--func", shq(funcName),
		"--gas-fee", "1gnot", "--gas-wanted", "2000000"];
	vals.forEach(function(arg) {
		args.push("--args");
		args.push(shq(arg));
	});
	var command = args.join(" ");
	command = command+" > unsigned.tx";
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// command 2: sign tx.
	var args = ["gnokey", "sign", "KEYNAME",
		"--txpath", "unsigned.tx", "--chainid", "testchain",
		"--number", "ACCOUNTNUMBER",
		"--sequence", "SEQUENCENUMBER"];
	var command = args.join(" ");
	command = command+" > signed.tx";
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// command 3: broadcast tx.
	var args = ["gnokey", "broadcast", "signed.tx"];
	var command = args.join(" ");
	command = command;
	shell.append(u("<span>").text(command)).append(u("<br>"));
}

function shq(s) {
	var s2 = String(s).replace(/\t/g, '\\t');
	var s2 = String(s2).replace(/\n/g, '\\n');
	var s2 = String(s2).replace(/([#!"$&'()*,:;<=>?@\[\\\]^`{|}])/g, '\\$1');
	return '"'+s2+'"';
};
