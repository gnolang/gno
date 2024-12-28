package dao_proposal_single

import (
	"strconv"
	"strings"

	"gno.land/p/demo/json"
	"gno.land/p/teritori/jsonutil"
)

// ported from https://github.com/DA0-DA0/dao-contracts/blob/7776858e780f1ce9f038a3b06cce341dd41d2189/packages/dao-voting/src/threshold.rs

type PercentageThreshold interface {
	String() string
	Clone() PercentageThreshold
	ToJSON() *json.Node
}

func PercentageThresholdFromJSON(ast *json.Node) PercentageThreshold {
	variant, valueNode := jsonutil.MustUnion(ast)
	switch variant {
	case "majority":
		return &PercentageThresholdMajority{}

	case "percent":
		p := PercentageThresholdPercent(0)
		p.FromJSON(valueNode)
		return &p
	}

	panic("unknown PercentageThreshold variant")
}

type PercentageThresholdMajority struct{}

func (p *PercentageThresholdMajority) String() string {
	return "Majority"
}

func (p *PercentageThresholdMajority) Clone() PercentageThreshold {
	return &PercentageThresholdMajority{}
}

func (p *PercentageThresholdMajority) ToJSON() *json.Node {
	return jsonutil.UnionNode("majority", jsonutil.EmptyObjectNode())
}

type PercentageThresholdPercent uint16 // 4 decimals fixed point

func (p *PercentageThresholdPercent) String() string {
	sb := strings.Builder{}
	sb.WriteString(strconv.FormatUint(uint64(*p)/100, 10))
	decPart := uint64(*p) % 100
	if decPart != 0 {
		sb.WriteRune('.')
		sb.WriteString(strconv.FormatUint(decPart, 10))
	}

	sb.WriteRune('%')
	return sb.String()
}

func (p *PercentageThresholdPercent) FromJSON(ast *json.Node) {
	val := ast.MustNumeric() // FIXME: don't pass by float64
	*p = PercentageThresholdPercent(val)
}

func (p *PercentageThresholdPercent) Clone() PercentageThreshold {
	c := *p
	return &c
}

func (p *PercentageThresholdPercent) ToJSON() *json.Node {
	return jsonutil.UnionNode("percent", json.NumberNode("", float64(*p)))
}

type Threshold interface {
	Clone() Threshold
	ToJSON() *json.Node
}

func ThresholdFromJSON(ast *json.Node) Threshold {
	variant, valueNode := jsonutil.MustUnion(ast)
	switch variant {
	case "absolutePercentage":
		panic("not implemented")

	case "thresholdQuorum":
		t := &ThresholdThresholdQuorum{}
		t.FromJSON(valueNode)
		return t

	case "absoluteCount":
		val := valueNode.MustNumeric()
		ac := ThresholdAbsoluteCount(val)
		return &ac
	}

	panic("unknown Threshold variant")
}

type ThresholdAbsolutePercentage struct {
	Value PercentageThreshold
}

func (t ThresholdAbsolutePercentage) Clone() Threshold {
	c := t.Value.Clone()
	return &ThresholdAbsolutePercentage{Value: c}
}

func (t ThresholdAbsolutePercentage) ToJSON() *json.Node {
	return jsonutil.UnionNode("absolutePercentage", t.Value.ToJSON())
}

type ThresholdThresholdQuorum struct {
	Threshold PercentageThreshold
	Quorum    PercentageThreshold
}

func (t *ThresholdThresholdQuorum) Clone() Threshold {
	return &ThresholdThresholdQuorum{
		Threshold: t.Threshold.Clone(),
		Quorum:    t.Quorum.Clone(),
	}
}

// TODO: make union members json utilities bijective

func (t *ThresholdThresholdQuorum) FromJSON(ast *json.Node) {
	obj := ast.MustObject()
	t.Threshold = PercentageThresholdFromJSON(obj["threshold"])
	t.Quorum = PercentageThresholdFromJSON(obj["quorum"])
}

func (t *ThresholdThresholdQuorum) ToJSON() *json.Node {
	return jsonutil.UnionNode("thresholdQuorum", json.ObjectNode("", map[string]*json.Node{
		"threshold": t.Threshold.ToJSON(),
		"quorum":    t.Quorum.ToJSON(),
	}))
}

type ThresholdAbsoluteCount uint64

func (t *ThresholdAbsoluteCount) Clone() Threshold {
	val := *t
	return &val
}

func (t *ThresholdAbsoluteCount) ToJSON() *json.Node {
	return jsonutil.UnionNode("absoluteCount", jsonutil.Uint64Node(uint64(*t)))
}
