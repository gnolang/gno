package gor

import (
	"std"
)

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
    GithubUsername std.Address `json"github_username"`
	prIDs []int `json:"pr_ids"`
}
