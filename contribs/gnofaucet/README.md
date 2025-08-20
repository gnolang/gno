# Start a local faucet

## Step1:

Make sure you have started gnoland
    
    ../../gno.land/build/gnoland start -lazy -skip-genesis-sig-verification

## Step2:

Start the faucet. This repository provides middleware for integrating GitHub OAuth authentication or reCAPTCHA verification into the Gno.land faucet. This ensures security by preventing abuse while enabling users to claim tokens securely.
#### Running Recapcha protected faucet:

    ./build/gnofaucet serve captcha  -chain-id dev -mnemonic "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast" --captcha-secret=<RECAPTCHA_SECRET>
    
| Flag                 | Type      | Default       | Description |
|----------------------|-----------|--------------|-------------|
| `--captcha-secret`  | `string`  | `""` (empty) | reCAPTCHA secret key. If empty, an errCaptchaMissing error is returned. |


#### Running Github Oauth protected faucet:

    ./build/gnofaucet serve github -chain-id dev -mnemonic "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast" --github-client-id=<CLIENT_ID> --cooldown-period=24h --max-claimable-limit=100000000 (100 gnot)

| Flag                  | Type       | Default       | Description |
|-----------------------|------------|--------------|-------------|
| `--github-client-id`  | `string`   | `""` (empty) | GitHub client ID for OAuth authentication. |
| `--cooldown-period`   | `duration` | `24h`        | Minimum required time between consecutive claims by the same user. |
| `--max-claimable-limit` | `int64`  | `0`          | Maximum number of tokens a user can claim over their lifetime. Zero means no limit |

By default, the faucet sends out 10,000,000ugnot (10gnot) per request. 

## Step3:

Make sure you have started website

    ../../gno.land/build/gnoweb

Request testing tokens from following URL, Have fun!

    http://localhost:8888/faucet