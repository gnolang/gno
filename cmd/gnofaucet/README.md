# Start a local faucet 

## Step1: Import test1 key
If you have imported the test1 key skip to Step2

    ./build/gnokey add test1 --recover

At prompt, input and confirm your password to protect the imported private key.

Copy and paste the following mnemonic.

    source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast


## Step2:

    ./build/faucet serve test1 --chain-id dev 

By default, the faucet sends out 1,000,000ugnot (1gnot) per request. If this is your local faucet, you can be a bit generous to yourself with --send flag. With the following, the faucet will give you 500gnot per request.

    ./build/faucet serve test1 --chain-id dev --send 5000000000ugnot
    
    
    
## Step3:

Make sure you have started gnoland and website 

    ./build/gnoland
   
    ./build/website
  
Request testing tokens from following URL, Have fun!

    http://localhost:8888/faucet

