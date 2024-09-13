async function main() {
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

  async function fetchAccount() {
    try {
      const response = await adena.GetAccount();
      if (response.code === 0 && response.status === "success") {
        const address = response.data.address;
        u("#my_address").first().value = address;
        setMyAddress(address);
        u("div.func_spec").each(function (node, i) {
          updateCommand(u(node));
        });
      }
    } catch (error) {
      console.error("Error fetching account:", error);
    }
  }

  //check url params have wallet_address or not
  var urlParams = new URLSearchParams(window.location.search);
  var connectAdena = urlParams.get('connect-adena');
  if (connectAdena) {
    if (!window.adena) {
      //open adena.app in a new tab if the adena object is not found
      window.open("https://adena.app/", "_blank");
    } else {
      //the sample code below displays a method provided by Adena that initiates a connection
      await adena.AddEstablish("Adena");
      await fetchAccount();
    }
  }
  //check url params have adena_message or not
  var adenaMessage = urlParams.get('adena_message');
  if (adenaMessage) {
    //get current url
    var currentUrl = window.location.href;
    //remove all url params
    currentUrl = currentUrl.split("?")[0];
    //redirect to current url
    window.location.href = currentUrl;
    var message = JSON.parse(atob(adenaMessage));
    await adena.DoContract(message);
  }

  // special case: when address changes.
  u("#my_address").on("input", function(e) {
    var value = u("#my_address").first().value;
    setMyAddress(value)
    u("div.func_spec").each(function(node, i) {
      updateCommand(u(node));
    });
  });
};

function setMyAddress(addr) {
  localStorage.setItem("my_address", addr);
}

function getMyAddress() {
  var myAddr = localStorage.getItem("my_address");
  if (!myAddr) {
    return "";
  }
  return myAddr;
}

// x: the u("div.func_spec") element.
function updateCommand(x) {
  var realmPath = u("#data").data("realm-path");
  var remote = u("#data").data("remote");
  var chainid = u("#data").data("chainid");
  var funcName = x.data("func-name");
  var ins = x.find("table>tbody>tr.func_params input");
  var vals = [];
  ins.each(function(input) {
    vals.push(input.value);
  });
  var myAddr = getMyAddress() || "ADDRESS";
  var shell = x.find(".shell_command");
  shell.empty();

  // command Z: all in one.
  shell.append(u("<span>").text("### INSECURE BUT QUICK ###")).append(u("<br>"));
  var args = ["gnokey", "maketx", "call",
    "-pkgpath", shq(realmPath), "-func", shq(funcName),
    "-gas-fee", "1000000ugnot", "-gas-wanted", "2000000",
    "-send", shq(""),
    "-broadcast", "-chainid", shq(chainid)];
  vals.forEach(function(arg) {
    args.push("-args");
    args.push(shq(arg));
  });
  args.push("-remote", shq(remote));
  args.push(myAddr);
  var command = args.join(" ");
  shell.append(u("<span>").text(command)).append(u("<br>")).append(u("<br>"));

  // or...
  shell.append(u("<span>").text("### FULL SECURITY WITH AIRGAP ###")).append(u("<br>"));

  // command 0: query account info.
  var args = ["gnokey", "query", "-remote", shq(remote), "auth/accounts/" + myAddr];
  var command = args.join(" ");
  shell.append(u("<span>").text(command)).append(u("<br>"));

  // command 1: construct tx.
  var args = ["gnokey", "maketx", "call",
    "-pkgpath", shq(realmPath), "-func", shq(funcName),
    "-gas-fee", "1000000ugnot", "-gas-wanted", "2000000",
    "-send", shq("")];
  vals.forEach(function(arg) {
    args.push("-args");
    args.push(shq(arg));
  });
  args.push(myAddr)
  var command = args.join(" ");
  command = command + " > call.tx";
  shell.append(u("<span>").text(command)).append(u("<br>"));

  // command 2: sign tx.
  var args = ["gnokey", "sign",
    "-tx-path", "call.tx", "-chainid", shq(chainid),
    "-account-number", "ACCOUNTNUMBER",
    "-account-sequence", "SEQUENCENUMBER", myAddr];
  var command = args.join(" ");
  shell.append(u("<span>").text(command)).append(u("<br>"));

  // command 3: broadcast tx.
  var args = ["gnokey", "broadcast", "-remote", shq(remote), "call.tx"];
  var command = args.join(" ");
  command = command;
  shell.append(u("<span>").text(command)).append(u("<br>")).append(u("<br>"));;

  // command 4: Sign and broadcast by Adena
  var adenaArgs = [];
  vals.forEach(function (arg) {
    adenaArgs.push(arg);
  });
  const message = {
    messages: [{
      type: "/vm.m_call",
      value: {
        caller: myAddr,
        send: "",
        pkg_path: realmPath,
        func: funcName,
        args: adenaArgs
      }
    }],
    gasFee: 1000000,
    gasWanted: 2000000
  };
  //convert message to base64
  const messageBase64 = btoa(JSON.stringify(message));
  var currentUrl = window.location.href;
  adena_href = currentUrl + "&adena_message=" + messageBase64;
  shell.append(u("<span>").text("### SIGN AND BROACAST BY ADENA")).append(u("<br>"));
  shell.append(u("<a href=" + adena_href + ">").text("Do Contract")).append(u("<br>"));
}

// Jae: why isn't this a library somewhere?
function shq(s) {
  var s2 = String(s).replace(/\t/g, '\\t');
  var s2 = String(s2).replace(/\n/g, '\\n');
  var s2 = String(s2).replace(/([$'"`\\!])/g, '\\$1');
  return '"' + s2 + '"';
};
