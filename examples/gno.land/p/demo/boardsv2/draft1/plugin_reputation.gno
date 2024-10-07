package boardsv2

// TODO: Move to a plugins folder
// TODO: Plugins can also have global stores (plugins can be local to a post or global)

func NewReputationPlugin( /* options */ ) ReputationPlugin {

}

type ReputationStorage struct {
	Upvotes            uint
	Downbotes          uint
	ListOfWhoVotedWhat avl.Tree // string(std.Address) -> ?? (TODO: define)
}

type ReputationPlugin struct {
	AllowedLevels []int
}

func (p ReputationPlugin) Name() string {
	return "reputation_plugin"
}

func (p *ReputationPlugin) Votes(p *Post) uint32 {
}

func (p *ReputationPlugin) Upvote(p *Post) {
	// TODO: Assert that post has the allowed level
	st := p.GetPluginStore(p.Name()).(*Storage)

	// TODO: modify global state
	// TODO: modify local state
}

func (p *ReputationPlugin) Downvote(p *Post) {
	// TODO: Assert that post has the allowed level

}
