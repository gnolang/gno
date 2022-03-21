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
	var args = ["gnokey", "maketx", "call", "<KEYNAME>", 
		"--pkgpath", realmPath, "--func", funcName,
		"--gas-fee", "1gnot", "--gas-wanted", "2000000"];
	vals.forEach(function(arg) {
		args.push("--arg");
		args.push(arg);
	});
	var command = shellescape(args);
	command = "> ./"+command+" > unsigned.tx";
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// command 2: sign tx.
	var args = ["gnokey", "sign", "<KEYNAME>",
		"--txpath", "unsigned.tx", "--chainid", "testchain",
		"--number", "<ACCOUNTNUMBER>",
		"--sequence", "<SEQUENCENUMBER>"];
	var command = shellescape(args);
	command = "> ./"+command+" > signed.tx";
	shell.append(u("<span>").text(command)).append(u("<br>"));

	// command 3: broadcast tx.
	var args = ["gnokey", "broadcast", "signed.tx"];
	var command = shellescape(args);
	command = "> ./"+command;
	shell.append(u("<span>").text(command)).append(u("<br>"));
}

// From https://github.com/xxorax/node-shell-escape/blob/master/shell-escape.js
// MIT license.
//
// return a shell compatible format
function shellescape(a) {
	var ret = [];

	a.forEach(function(s) {
		if (/[^A-Za-z0-9_\/:=-]/.test(s)) {
			s = "'"+s.replace(/'/g,"'\\''")+"'";
			s = s.replace(/^(?:'')+/g, '') // unduplicate single-quote at the beginning
				.replace(/\\'''/g, "\\'" ); // remove non-escaped single-quote if there are enclosed between 2 escaped
		}
		if (s == "") {
			s = "''";
		}
		ret.push(s);
	});

	return ret.join(' ');
}
