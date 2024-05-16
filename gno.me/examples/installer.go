package examples

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.me/gno"
)

func CreateInstallerApp(vm gno.VM) error {
	appCode := fmt.Sprintf(appDefinition, "`"+renderContents+"`")
	return vm.Create(context.Background(), appCode, false)
}

const appDefinition = `
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
	<title>Create App</title>
	<script>
		function submitForm() {
			var formData = {
				code: document.getElementById("code").value
			};

			fetch('http://localhost:4591/system/create', {
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
	<h2>Create App</h2>
	<form id="myForm">
		<label for="code">Code:</label><br>
		<textarea id="code" name="code" rows="50" cols="150"></textarea><br><br>
		<input type="button" value="Submit" onclick="submitForm()">
	</form>
</body>

</html>
`
