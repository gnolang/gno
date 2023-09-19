package dao_proposal_single

import (
	"strconv"

	"gno.land/p/demo/teritori/ujson"
)

// ported from https://github.com/DA0-DA0/dao-contracts/blob/7776858e780f1ce9f038a3b06cce341dd41d2189/packages/dao-voting/src/threshold.rs

type PercentageThreshold interface {
	String() string
	Clone() PercentageThreshold
	ToJSON() string
}

func parseUnion(ast *ujson.JSONASTNode, kv []*ujson.ParseKV) interface{} {
	if ast.Kind != ujson.JSONKindObject {
		panic("union is not an object")
	}
	if len(ast.ObjectChildren) != 1 {
		panic("union object does not have exactly one field")
	}
	k, node := ast.ObjectChildren[0].Key, ast.ObjectChildren[0].Value
	for _, kv := range kv {
		if kv.Key == k {
			node.ParseAny(kv.Value)
			return kv.Value
		}
	}
	panic("unknown union type") // TODO: expected one of ...
}

func formatUnionMember(name string, val interface{}, raw bool) string {
	return ujson.FormatObject([]ujson.FormatKV{
		{Key: name, Value: val, Raw: raw},
	})
}

func PercentageThresholdFromJSON(ast *ujson.JSONASTNode) PercentageThreshold {
	p := PercentageThresholdPercent(0)
	return parseUnion(ast, []*ujson.ParseKV{
		{"majority", &PercentageThresholdMajority{}},
		{"percent", &p},
	}).(PercentageThreshold)
}

type PercentageThresholdMajority struct{}

func (p *PercentageThresholdMajority) String() string {
	return "Majority"
}

func (p *PercentageThresholdMajority) Clone() PercentageThreshold {
	return &PercentageThresholdMajority{}
}

func (p *PercentageThresholdMajority) ToJSON() string {
	return formatUnionMember("majority", "{}", true)
}

type PercentageThresholdPercent uint16 // 4 decimals fixed point

func (p *PercentageThresholdPercent) String() string {
	s := strconv.FormatUint(uint64(*p)/100, 10)
	decPart := uint64(*p) % 100
	if decPart != 0 {
		s += "." + strconv.FormatUint(decPart, 10)
	}
	return s + "%"
}

func (p *PercentageThresholdPercent) FromJSON(ast *ujson.JSONASTNode) {
	var val uint32
	ujson.ParseAny(ast.Value, &val)
	*p = PercentageThresholdPercent(val)
}

func (p *PercentageThresholdPercent) Clone() PercentageThreshold {
	c := *p
	return &c
}

func (p *PercentageThresholdPercent) ToJSON() string {
	return formatUnionMember("percent", uint64(*p), false)
}

type Threshold interface {
	Clone() Threshold
	ToJSON() string
}

func ThresholdFromJSON(ast *ujson.JSONASTNode) Threshold {
	ac := ThresholdAbsoluteCount(0)
	return parseUnion(ast, []*ujson.ParseKV{
		// TODO: {Key: "absolutePercentage"},
		{Key: "thresholdQuorum", Value: &ThresholdThresholdQuorum{}},
		{Key: "absoluteCount", Value: &ac},
	}).(Threshold)
}

type ThresholdAbsolutePercentage PercentageThreshold

func (t *ThresholdAbsolutePercentage) Clone() Threshold {
	c := (*t).(PercentageThreshold).Clone()
	return (*ThresholdAbsolutePercentage)(c)
}

func (t *ThresholdAbsolutePercentage) ToJSON() string {
	return formatUnionMember("absolutePercentage", (t).(*PercentageThreshold), false)
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

func (t *ThresholdThresholdQuorum) FromJSON(ast *ujson.JSONASTNode) {
	ast.ParseObject([]*ujson.ParseKV{
		{Key: "threshold", CustomParser: func(ast *ujson.JSONASTNode) {
			t.Threshold = PercentageThresholdFromJSON(ast)
		}},
		{Key: "quorum", CustomParser: func(ast *ujson.JSONASTNode) {
			t.Quorum = PercentageThresholdFromJSON(ast)
		}},
	})
}

func (t *ThresholdThresholdQuorum) ToJSON() string {
	return formatUnionMember("thresholdQuorum", ujson.FormatObject([]ujson.FormatKV{
		{Key: "threshold", Value: t.Threshold},
		{Key: "quorum", Value: t.Quorum},
	}), true)
}

type ThresholdAbsoluteCount uint64

func (t *ThresholdAbsoluteCount) Clone() Threshold {
	return &ThresholdAbsoluteCount(*t)
}

func (t *ThresholdAbsoluteCount) ToJSON() string {
	return formatUnionMember("absoluteCount", uint64(*t), false)
}
