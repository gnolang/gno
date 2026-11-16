package json

type NodeBuilder struct {
	node *Node
}

func Builder() *NodeBuilder {
	return &NodeBuilder{node: ObjectNode("", nil)}
}

func (b *NodeBuilder) WriteString(key, value string) *NodeBuilder {
	b.node.AppendObject(key, StringNode("", value))
	return b
}

func (b *NodeBuilder) WriteNumber(key string, value float64) *NodeBuilder {
	b.node.AppendObject(key, NumberNode("", value))
	return b
}

func (b *NodeBuilder) WriteBool(key string, value bool) *NodeBuilder {
	b.node.AppendObject(key, BoolNode("", value))
	return b
}

func (b *NodeBuilder) WriteNull(key string) *NodeBuilder {
	b.node.AppendObject(key, NullNode(""))
	return b
}

func (b *NodeBuilder) WriteObject(key string, fn func(*NodeBuilder)) *NodeBuilder {
	nestedBuilder := &NodeBuilder{node: ObjectNode("", nil)}
	fn(nestedBuilder)
	b.node.AppendObject(key, nestedBuilder.node)
	return b
}

func (b *NodeBuilder) WriteArray(key string, fn func(*ArrayBuilder)) *NodeBuilder {
	arrayBuilder := &ArrayBuilder{nodes: []*Node{}}
	fn(arrayBuilder)
	b.node.AppendObject(key, ArrayNode("", arrayBuilder.nodes))
	return b
}

func (b *NodeBuilder) Node() *Node {
	return b.node
}

type ArrayBuilder struct {
	nodes []*Node
}

func (ab *ArrayBuilder) WriteString(value string) *ArrayBuilder {
	ab.nodes = append(ab.nodes, StringNode("", value))
	return ab
}

func (ab *ArrayBuilder) WriteNumber(value float64) *ArrayBuilder {
	ab.nodes = append(ab.nodes, NumberNode("", value))
	return ab
}

func (ab *ArrayBuilder) WriteInt(value int) *ArrayBuilder {
	return ab.WriteNumber(float64(value))
}

func (ab *ArrayBuilder) WriteBool(value bool) *ArrayBuilder {
	ab.nodes = append(ab.nodes, BoolNode("", value))
	return ab
}

func (ab *ArrayBuilder) WriteNull() *ArrayBuilder {
	ab.nodes = append(ab.nodes, NullNode(""))
	return ab
}

func (ab *ArrayBuilder) WriteObject(fn func(*NodeBuilder)) *ArrayBuilder {
	nestedBuilder := &NodeBuilder{node: ObjectNode("", nil)}
	fn(nestedBuilder)
	ab.nodes = append(ab.nodes, nestedBuilder.node)
	return ab
}

func (ab *ArrayBuilder) WriteArray(fn func(*ArrayBuilder)) *ArrayBuilder {
	nestedArrayBuilder := &ArrayBuilder{nodes: []*Node{}}
	fn(nestedArrayBuilder)
	ab.nodes = append(ab.nodes, ArrayNode("", nestedArrayBuilder.nodes))
	return ab
}
