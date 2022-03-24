

## Install

      git clone https://github.com/piux2/gnobounty5

      cd gnobounty5

      make all

My code is based on

The codebase committed on Dec 9, 2021
https://github.com/gnolang/gno/tree/5a1ea776cac472a42e3b0ecf4d32ebc1ede289f9


## Testing data

  #### Primary Key:

  name: test1

  mnemonic:

        source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast

  passphrase: test 1

  generated address and pubkey

        test1 (local) - addr: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
        pub: gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj

  ####  Backup Key

  name: test1

  mnemonic:

    curious syrup memory cabbage razor emotion ketchup best alley cotton enjoy nature furnace shallow donor oval tornado razor clock roof pave enroll solar wrist

  generated multisig address and pubkey

    test1 (backup local- multisig address) - addr: g16ptpek560p53qdmeja7vm2crc0gpgtqyzfuthv
    multisig pub:
    [0]gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj
    [1]gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp7xtkykvttvcxnz9n74hfd8t4tav3t7l33p5trvyeuxd3ea8d95vhp767p


## Instructions

#### Start from here if you have not created primary key yet, otherwise skip to the next step
Note: Enter words within < >. Do not enter brackets

    ./build/gnokeybk add test1 --recover
    Enter a passphrase to encrypt your key to disk: <test1>

    Enter your bip39 mnemonic
    <source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast>

    test1 (local) - addr: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 pub: gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj, path: <nil>


#### Start from here if you have created primary key yet.
Back up your primary key to a seperate backup keybase

    ./build/gnokeybk bkkey test1
    Enter a passphrase to encrypt your key to disk: <test1>

    Enter your backup bip39 mnemonic, which should be different from you primary mnemonic, or hit enter to generate a new one
    <curious syrup memory cabbage razor emotion ketchup best alley cotton enjoy nature furnace shallow donor oval tornado razor clock roof pave enroll solar wrist>

    Backup key  is created for primary key address
    g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5

    Backup key's multisig address is
    g16ptpek560p53qdmeja7vm2crc0gpgtqyzfuthv

### check and list primary key and backup key. both share the same name

    ./build/gnokeybk listbk

    Keybase primary
    0. test1 (local) - addr: g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 pub: gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj, path: <nil>


    ---------------------------
    Keybase backup
    0. test1 (local) - addr: g16ptpek560p53qdmeja7vm2crc0gpgtqyzfuthv pub: gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj | gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp7xtkykvttvcxnz9n74hfd8t4tav3t7l33p5trvyeuxd3ea8d95vhp767p | , path: <nil>



### Sign transactions with the backup key

Launch the gnoland chain in a separate terminal

./build/gnoland

Check both keys are available on-chain. these are preconfigured accounts in genesis.
In the real case, you will have to send the token to your backup key address, which is a multisig address.

Before you sign and broadcasted messages. these two accounts on chain do not have pub keys published on the chain yet.

      ./build/gnokeybk query "auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

      height: 0
      data: {
        "BaseAccount": {
          "address": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
          "coins": "1000000gnot",
          "public_key": null,
          "account_number": "0",
          "sequence": "0"
        }
      }

      ./build/gnokeybk query "auth/accounts/g16ptpek560p53qdmeja7vm2crc0gpgtqyzfuthv"

      height: 0
      data: {
        "BaseAccount": {
          "address": "g16ptpek560p53qdmeja7vm2crc0gpgtqyzfuthv",
          "coins": "1000000gnot",
          "public_key": null,
          "account_number": "1",
          "sequence": "0"
        }
      }



Now let's use the backup key to sign and broadcast the signed transaction
This transaction is created by test1 following the examples in https://github.com/gnolang/gno/tree/master/examples/gno.land/r/boards

Let's sign it with the backup key. The account number is set in genesis. We need to increment sequence number each time we sign a transaction. The signer's address will be replaced by the backup key multisig address.

      ./build/gnokeybk signbk test1 --txpath addpkg.avl.unsigned.json --chainid "testchain" --number 1 --sequence 0 > addpkg.avl.signed.json

      Enter password.
      <test1>

### broadcast transactions

    ./build/gnokeybk broadcast addpkg.avl.signed.json

    $ OK!

The transaction is successfully broadcasted and accepted by the chain.

We query the backup key account on chain. The multisig pub key is published.


      .build/gnokeybk query "auth/accounts/g16ptpek560p53qdmeja7vm2crc0gpgtqyzfuthv"

      height: 0
      data: {
        "BaseAccount": {
          "address": "g16ptpek560p53qdmeja7vm2crc0gpgtqyzfuthv",
          "coins": "999898gnot",
          "public_key": {
            "@type": "/tm.PubKeyMultisig",
            "threshold": "2",
            "pubkeys": [
              {
                "@type": "/tm.PubKeySecp256k1",
                "value": "A+FhNtsXHjLfSJk1lB8FbiL4mGPjc50Kt81J7EKDnJ2y"
              },
              {
                "@type": "/tm.PubKeyEd25519",
                "value": "+MuxLMWtmDTEWfq3S066r6yK/fjENFjYTPDNjnp2low="
              }
            ]
          },
          "account_number": "1",
          "sequence": "1"
        }
      }


DONE!
