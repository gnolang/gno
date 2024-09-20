# Start a local faucet

## Step1:

Make sure you have started gnoland
    
    ../../gno.land/build/gnoland start -lazy

## Step2:

Start the faucet.

    ./build/gnofaucet serve -chain-id dev -mnemonic "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"

By default, the faucet sends out 10,000,000ugnot (10gnot) per request. 

## Step3:

Make sure you have started website

    ../../gno.land/build/gnoweb

Request testing tokens from following URL, Have fun!

    http://localhost:8888/faucet