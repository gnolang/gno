// names_bootstrap is a temporary namespace verifier deployed during the namespace v2 setup.
// It extends PA-namespace logic to allow the deployer to deploy into the sys namespace,
// enabling deployment of r/sys/names/v2 while enforcement is active.
// Replaced by r/sys/names/v2 immediately after it is deployed.
package names

// PATCH:DEPLOYER_ADDR — replaced by enable-basic-namespaces.sh
const bootstrapDeployer = "__DEPLOYER_ADDR__"

func IsAuthorizedAddressForNamespace(address_XXX address, namespace string) bool {
	if namespace == "" || !address_XXX.IsValid() {
		return false
	}
	if address_XXX.String() == namespace {
		return true // PA namespace: always allowed
	}
	// Allow bootstrapDeployer to deploy into sys (for r/sys/names/v2 only)
	return address_XXX.String() == bootstrapDeployer && namespace == "sys"
}
