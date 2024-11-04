package pluginpoll

const Name = "post-poll"

type (
	Plugin struct {
		posts post.Store
	}

	Poll struct {
		Question string
		Options  []string
		Votes    []struct {
			Address std.Adress
			Option  string
		}
		Tags []string
	}
)

func New(st post.Store) Plugin {
	return Plugin{
		posts: st,
	}
}

func (p Plugin) Name() string {
	return Name
}

func (p Plugin) Render() string {
	return ""
}

func (p Plugin) CreatePoll(id string, v Poll) *post.Post {
	pst := &post.Post{
		ID:    id,
		Level: LevelPost,
	}
	p.SetPoll(pst, v)
	return pst
}

func (p Plugin) SetPoll(pst *post.Post, v Poll) (updated bool) {
	pst.PluginStore[Name] = v
	return p.posts.Set(pst.ID, pst)
}
