package ui

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.me/gno"
)

func AddInstallerRealm(vm gno.VM) error {
	realm := fmt.Sprintf(realmDefinition, "`"+renderContents+"`")
	addPkg := gno.NewMsgAddPackage("installer", realm)
	return vm.AddPackage(context.Background(), addPkg)
}

const realmDefinition = `
package installer

func Render(_ string) string {
	return %s
}
`

const renderContents = `
<!DOCTYPE html>
<html lang="en">

<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Simple Form</title>
	<script>
		function submitForm() {
			var formData = {
				name: document.getElementById("name").value,
				code: document.getElementById("code").value
			};

			fetch('http://localhost:4591/system/install', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Access-Control-Allow-Origin': '*'
				},
				body: JSON.stringify(formData)
			})
				.then(response => {
					console.log(response);
				})
				.catch(error => {
					console.error('Error:', error);
				});
		}
	</script>
</head>

<body>
	<h2>Submit Form</h2>
	<form id="myForm">
		<label for="name">Name:</label><br>
		<input type="text" id="name" name="name"><br>
		<label for="code">Code:</label><br>
		<textarea id="code" name="code" rows="4" cols="50"></textarea><br><br>
		<input type="button" value="Submit" onclick="submitForm()">
	</form>
</body>

</html>
`
