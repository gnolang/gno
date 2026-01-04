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

    ./build/gnofaucet serve github -chain-id dev -mnemonic "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast" --cooldown-period=24h --max-claimable-limit=100000000 (100 gnot)

| Flag                    | Type       | Default      | Description |
|-------------------------|------------|--------------|-------------|
| `--cooldown-period`     | `duration` | `24h`        | Minimum required time between consecutive claims by the same user. |
| `--max-claimable-limit` | `int64`    | `0`          | Maximum number of tokens a user can claim over their lifetime. Zero means no limit |

By default, the faucet sends out 10,000,000ugnot (10gnot) per request. 

#### Running Github Fetcher

To run the GitHub fetcher, which is a utility for fetching and storing GitHub user scores (such as username, commits, issues and PRs counts), use the following command:

    ./build/gnofaucet fetcher github --github-client-id=<CLIENT_ID> --github-client-secret=<CLIENT_SECRET> --github-username=<USERNAME>

| Flag                    | Type       | Default | Description |
|-------------------------|------------|---------|-------------|
| `--fetch-interval `     | `duration` | `"20s"` | Interval used to fetch new events from Github repositories. |

This command will fetch and save information on redis about the specified GitHub user, which will be used for verifying eligibility for faucet rewards. Make sure to fill all requested environment variables like Github Apps and Redis ones.

## Step3:

Make sure you have started website

    ../../gno.land/build/gnoweb

Request testing tokens from following URL, Have fun!

    http://localhost:8888/faucet