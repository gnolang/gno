package apps

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/gno.me/gno"
)

func CreateInstaller(vm gno.VM) error {
	renderContents := fmt.Sprintf("`%s` + port.Number() + `%s`", prePortContents, postPortContents)
	appCode := fmt.Sprintf(appDefinition, renderContents)
	_, err := vm.Create(context.Background(), appCode, false, false)
	return err
}

const appDefinition = `
package installer

import "gno.land/r/port"

func Render(_ string) string {
	return %s
}
`

const prePortContents = `
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

			fetch('http://localhost:`

const postPortContents = `/system/create', {
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
	<h2>Create App</h2>
	<form id="myForm">
		<label for="code">Code:</label><br>
		<textarea id="code" name="code" rows="30" cols="100"></textarea><br><br>
		<input type="button" value="Submit" onclick="submitForm()">
	</form>
</body>

</html>
`
