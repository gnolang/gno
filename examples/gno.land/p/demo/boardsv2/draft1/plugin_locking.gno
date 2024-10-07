package boardsv2

func NewLockingPlugin( /* options */ ) LockingPlugin {
}

type LockingStorage struct {
	IsLocked bool
}

type LockingPlugin struct {
}

func (p *LockingPlugin) Lock(p *Post) {
	// TODO: Check local storare to see if it's lock otherwise change storage
	p.GetPluginStore(p.Name()).(LockingStorage).IsLocked = true
	// TODO: We need to update (Save) the post everytime there is a change.
}
func (p *LockingPlugin) Unlock(p *Post)       {}
func (p LockingPlugin) IsLocked(p *Post) bool {}
