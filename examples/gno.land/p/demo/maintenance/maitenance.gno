package maintenance

import (
	"std"
	"time"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
)

func New(owner *ownable.Ownable) *Config {
	return nil
	/*return &Config{
		ownable.Ownable: owner,
		lastUpdate:      time.Now(),
		inMaintenance:   false,
	}*/
}

// Config is a safe object that can be exposed so that other contracts can read its content.
// It is also safe to not expose all the helpers methods from here on the final realm, but to just expose the pointer and expect the admin to use a maketx run action to perform the maintenance updates.
type Config struct {
	//*ownable.Ownable
	inMaintenance bool
	body          string    // body is an optional field which will be added on the Render function whatever the reason is. it could be used to share contact info.
	lastUpdate    time.Time // lastUpdate is set to the last date wether it's to declare a maintenance or resolve one
	reason        string    // reason is an optional field that is used when declaring a maintenance or resolving one that will be printed to the users in render and panics.
}

/*

func (c Config) InMaintenance() bool   { return c.inMaintenance }
func (c Config) Reason() string        { return c.reason }
func (c Config) LastUpdate() time.Time { return c.lastUpdate }
func (c Config) Body() string          { return c.body }

func (c Config) RenderMaintenance() string {
	if c.inMaintenance {
		return ""
	}
	// output := ufmt.String("In maintenance since %s\n\n", time.Since(c.lastUpdate), c.reason)
	// return output
	return "TODO"
}

func (c Config) AssertInMaintenance(comment string) {
	if c.inMaintenance {
		panic(c.RenderMaintenance())
	}
}

func (c *Config) DeclareMaintenance(reason string) error {
	if err := c.CallerIsOwner(); err != nil {
		return err
	}

	if !c.inMaintenance {
		c.lastUpdate = time.Now()
		c.inMaintenance = true
	}
	c.reason = reason // always update the reason
}

func (c *Config) ResolveMaintenance(reason string) error {
	if err := c.CallerIsOwner(); err != nil {
		return err
	}

	if c.inMaintenance {
		c.lastUpdate = time.Now()
		c.inMaintenance = false
	}
	c.reason = reason // always update the reason
}

func (c *Config) UpdateBody(newBody string) error {
	if err := c.CallerIsOwner(); err != nil {
		return err
	}
	c.body = newBody
}
*/
