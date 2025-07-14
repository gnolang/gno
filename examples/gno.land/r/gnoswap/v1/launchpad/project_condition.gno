package launchpad

import (
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
)

// ProjectCondition represents a condition for a project.
//
// This struct contains the necessary data and methods to manage and distribute
// rewards for a specific project.
//
// Fields:
// - tokenPath (string): The path of the token associated with the project.
// - minimumAmount (int64): The minimum amount of the token required for the project.
type ProjectCondition struct {
	tokenPath     string
	minimumAmount int64
}

func (p *ProjectCondition) TokenPath() string {
	return p.tokenPath
}

func (p *ProjectCondition) MinimumAmount() int64 {
	return p.minimumAmount
}

func (p *ProjectCondition) IsAvailable() bool {
	return p.tokenPath != "" && p.minimumAmount > 0
}

func (p *ProjectCondition) CheckCondition(inputTokenPath string, inputAmount int64) error {
	if p.tokenPath != inputTokenPath {
		return makeErrorWithDetails(
			errInvalidData,
			ufmt.Sprintf("token path(%s) is not matched", inputTokenPath),
		)
	}

	if inputAmount < p.minimumAmount {
		return makeErrorWithDetails(
			errInvalidData,
			ufmt.Sprintf("input amount(%d) is less than minimum amount(%d)", inputAmount, p.minimumAmount),
		)
	}

	return nil
}

func NewProjectCondition(tokenPath string, minimumAmount int64) *ProjectCondition {
	return &ProjectCondition{
		tokenPath:     tokenPath,
		minimumAmount: minimumAmount,
	}
}

func NewProjectConditions(conditionTokens string, conditionAmounts string) []*ProjectCondition {
	if conditionTokens == "" && conditionAmounts == "" {
		return []*ProjectCondition{}
	}

	conditions := []*ProjectCondition{}

	tokenPaths := strings.Split(conditionTokens, stringSplitterPad)
	minimumAmounts := strings.Split(conditionAmounts, stringSplitterPad)

	for index, tokenPath := range tokenPaths {
		if index >= len(minimumAmounts) {
			panic(ufmt.Sprintf("condition amount(%s) is not matched with condition token(%s)", conditionAmounts, conditionTokens))
		}

		minimumAmount, err := strconv.ParseInt(minimumAmounts[index], 10, 64)
		if err != nil {
			panic(err)
		}

		conditions = append(conditions, NewProjectCondition(tokenPath, minimumAmount))
	}

	return conditions
}
