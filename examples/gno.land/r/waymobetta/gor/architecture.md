## Architecture

#### Flow
- CI triggers upon new PR submission
- Script collects raw PR data, preps and sends to Evaluation server
    - Runs AI models
    - Generates evaluation
    - Posts data to Realm
- Evaluation server posts results back to Github via PR ID

#### Components:
- [Registry (Github => Gno address)](#registry-github--gno-address)
- [Github CI](#github-ci)
- [Helper Script 1](#helper-script-1)
- [Evaluation Server](#evaluation-server)
- [Helper Script 2](#helper-script-2)
- [Realm](#realm)

#### Registry (Github => Gno Address)
- A registry is needed to associate a Github username with a Gno address (public key)
- This way we can associate the data from the PR to a Gno address and store this as a mapping within the Realm

#### Github CI
- Triggers based on new PR submission
- Collects raw data of PR

#### Helper Script 1
- Executes based on trigger of CI
- ETL of PR data to server housing AI models

#### Evaluation Server
- *External service (AWS; beefy EC2)
    * Perhaps not external due to needing to send the data back to Github CI and storing in the [realm](#realm)
    * Perhaps AI server actually does the posting to the Realm endpoint
- API
- Accepts PR data and returns results of model (ie, score/summary)

#### Helper Script 2
- Collects results of AI model and stores in [realm](#realm)
* This script is possibly stored on AI server which would also mean the private key of the [realm](#realm) admin is stored on server as it is needed for posting data

#### Realm
- Stores results of AI and associates it to user's address 
```go

// mappings

// github_username => gno_address
type GitGnoMapping map[string]std.Address

// gno_address => github_username
type GnoGitMapping map[std.Address]string

// PR struct
type PR struct {
    ID int `json:"id"`
    Evaluation int `json:"evaluation"`
    MaxEvaluation int `json:"max_evaluation"`
    Category string `json:"category"`
    SubCategory string `json:"sub_category"`
    Commits int `json:"commits"`
    Additions int `json:"additions"`
    Deletions int `json:"deletions"`
    TotalEffectiveLines int `json:"total_effective_lines"`
    AvgCharsPerLine int `json:"avg_chars_per_line"`
}

// Gor struct
type Gor struct {
    PR *PR `json:"pr"`
}
```
