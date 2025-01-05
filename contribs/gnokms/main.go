package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	vault "github.com/hashicorp/vault/api"
)

const (
	mnemonicEntropySize = 256
)

func main() {
	if len(os.Args) < 2 {
		panic("missing action")
	}

	// Initialize the Vault client
	vault, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		panic("unable to create vault client: " + err.Error())
	}

	// Set the Vault address and token (replace with your values)
	vault.SetAddress("http://127.0.0.1:8200")
	vault.SetToken("s.rPokUM38w9a3Y68rWX37OGam")

	kb := keys.NewInMemory()
	switch os.Args[1] {
	case "create":
		mnemonic, err := client.GenerateMnemonic(mnemonicEntropySize)
		if err != nil {
			panic("unable to generate mnemonic: " + err.Error())
		}

		info, err := kb.CreateAccount(mnemonic, mnemonic, "", "", 0, 0)
		if err != nil {
			panic("unable to create account: " + err.Error())
		}
		exported, err := kb.Export(mnemonic)
		if err != nil {
			panic("unable to export mnemonic: " + err.Error())
		}
		err = createUser(vault, info.GetAddress().String(), "securepassword123")
		if err != nil {
			panic("unable to create user: " + err.Error())
		}
		err = storeKey(vault, info.GetAddress().String(), exported)
		if err != nil {
			panic("unable to store key: " + err.Error())
		}
	case "import":
		if len(os.Args) < 3 {
			panic("missing account address")
		}
		secret, err := login(os.Args[2], "securepassword123")
		if err != nil {
			panic("unable to login: " + err.Error())
		}
		err = kb.ImportPrivKey(os.Args[2], secret.Data["armor"].(string), "", "")
		if err != nil {
			panic("unable to import key: " + err.Error())
		}
		kb.Sign(os.Args[2], "securepassword123", []byte("hello"))
		fmt.Println(secret.Data["armor"])
	}

	//kb.ImportPrivKey()
	//err = kb.Import(mnemonic, exported)
	//if err != nil {
	//	panic("unable to import mnemonic: " + err.Error())
	//}

	//err = login("g1lczd2nuzlp8ctfa8d79j82z2y5upul2er9ke3t", "securepassword123")
	//if err != nil {
	//	panic("unable to login: " + err.Error())
	//}
}

func createUser(client *vault.Client, address, password string) error {
	// Define the username and password for the new user
	policies := []string{"user-kv", "user-secrets"}

	// Create the user by writing to the auth/userpass/users/<username> endpoint
	userPath := fmt.Sprintf("auth/userpass/users/%s", address)
	_, err := client.Logical().Write(userPath, map[string]interface{}{
		"password": password,
		"policies": policies,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	fmt.Printf("User '%s' created successfully with policies: %v\n", address, policies)
	return nil
}

func login(address, password string) (*vault.Secret, error) {
	// Initialize the Vault client
	client, err := vault.NewClient(vault.DefaultConfig())
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}

	// Set the Vault server address
	client.SetAddress("http://127.0.0.1:8200") // Replace with your Vault address

	// Authenticate using the userpass auth method
	loginPath := fmt.Sprintf("auth/userpass/login/%s", address)
	authData := map[string]interface{}{
		"password": password,
	}

	// Perform the login request
	secret, err := client.Logical().Write(loginPath, authData)
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Println(secret.Auth.EntityID)

	// Extract the token from the response
	token := secret.Auth.ClientToken
	fmt.Printf("Login successful! Token: %s\n", token)

	// Set the token in the Vault client for future operations
	client.SetToken(token)

	// Example: Read a secret (replace with a valid path for the user)
	secretPath := fmt.Sprintf("kv/data/%s/mykey", address)
	kvSecret, err := client.Logical().Read(secretPath)
	if err != nil {
		log.Fatalf("Failed to read secret: %v", err)
	}
	//
	//// Print the secret data
	if kvSecret != nil && kvSecret.Data != nil {
		fmt.Printf("Secret data: %v\n", kvSecret.Data)
	} else {
		fmt.Println("No secret data found.")
	}
	return kvSecret, nil
}

func storeKey(client *vault.Client, address, armor string) error {
	secretPath := fmt.Sprintf("kv/data/%s/mykey", address)
	_, err := client.Logical().Write(secretPath, map[string]interface{}{
		"armor": armor,
	})
	if err != nil {
		return fmt.Errorf("failed to store key: %v", err)
	}
	return nil
}
