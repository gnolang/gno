package apps

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.me/gno"
)

func CreateRemoteInstaller(vm gno.VM) error {
	renderContents := fmt.Sprintf("`%s` + port.Number() + `%s`", remotePrePortContents, remotePostPortContents)
	appCode := fmt.Sprintf(remoteAppDefinition, renderContents)
	_, err := vm.Create(context.Background(), appCode, false, false)
	return err
}

const remoteAppDefinition = `
package remoteinstaller

import "gno.land/r/port"

func Render(_ string) string {
	return %s
}
`

const remotePrePortContents = `
<!DOCTYPE html>
<html lang="en">

<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Create App</title>
	<script>
		function submitForm() {
			var formData = {
				name: document.getElementById("name").value,
				address: document.getElementById("address").value
			};

			fetch('http://localhost:`

const remotePostPortContents = `/system/install-remote', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Access-Control-Allow-Origin': '*'
				},
				body: JSON.stringify(formData)
			})
				.then(response => {
					if (response.status == 201) {
						console.log(response);
						alert("Success!");
					} else {
						console.error('Error:', response);
						alert("Failure: " + response.statusText);
					}
				})
				.catch(error => {
					console.error('Error:', error);
					alert("Failure: " + error);
				});
		}
	</script>
</head>

<body>
	<h2>Install Remote App</h2>
	<form id="myForm">
		<label for="address">Address:</label><br>
		<input type="text" id="address" name="address"><br><br>
		<label for="name">Name:</label><br>
		<input type="text" id="name" name="name"><br><br>
		<input type="button" value="Submit" onclick="submitForm()">
	</form>
</body>

</html>
`
