package tamagotchi

import (
	"time"

	"gno.land/p/nt/ufmt"
)

// Tamagotchi structure
type Tamagotchi struct {
	name        string
	hunger      int
	happiness   int
	health      int
	age         int
	maxAge      int
	sleepy      int
	created     time.Time
	lastUpdated time.Time
}

func New(name string) *Tamagotchi {
	now := time.Now()
	return &Tamagotchi{
		name:        name,
		hunger:      50,
		happiness:   50,
		health:      50,
		maxAge:      100,
		lastUpdated: now,
		created:     now,
	}
}

func (t *Tamagotchi) Name() string {
	t.update()
	return t.name
}

func (t *Tamagotchi) Hunger() int {
	t.update()
	return t.hunger
}

func (t *Tamagotchi) Happiness() int {
	t.update()
	return t.happiness
}

func (t *Tamagotchi) Health() int {
	t.update()
	return t.health
}

func (t *Tamagotchi) Age() int {
	t.update()
	return t.age
}

func (t *Tamagotchi) Sleepy() int {
	t.update()
	return t.sleepy
}

// Feed method for Tamagotchi
func (t *Tamagotchi) Feed() {
	t.update()
	if t.dead() {
		return
	}
	t.hunger = bound(t.hunger-10, 0, 100)
}

// Play method for Tamagotchi
func (t *Tamagotchi) Play() {
	t.update()
	if t.dead() {
		return
	}
	t.happiness = bound(t.happiness+10, 0, 100)
}

// Heal method for Tamagotchi
func (t *Tamagotchi) Heal() {
	t.update()

	if t.dead() {
		return
	}
	t.health = bound(t.health+10, 0, 100)
}

func (t Tamagotchi) dead() bool { return t.health == 0 }

// Update applies changes based on the duration since the last update
func (t *Tamagotchi) update() {
	if t.dead() {
		return
	}

	now := time.Now()
	if t.lastUpdated == now {
		return
	}

	duration := now.Sub(t.lastUpdated)
	elapsedMins := int(duration.Minutes())

	t.hunger = bound(t.hunger+elapsedMins, 0, 100)
	t.happiness = bound(t.happiness-elapsedMins, 0, 100)
	t.health = bound(t.health-elapsedMins, 0, 100)
	t.sleepy = bound(t.sleepy+elapsedMins, 0, 100)

	// age is hours since created
	t.age = int(now.Sub(t.created).Hours())
	if t.age > t.maxAge {
		t.age = t.maxAge
		t.health = 0
	}
	if t.health == 0 {
		t.sleepy = 0
		t.happiness = 0
		t.hunger = 0
	}

	t.lastUpdated = now
}

// Face returns an ASCII art representation of the Tamagotchi's current state
func (t *Tamagotchi) Face() string {
	t.update()
	return t.face()
}

func (t *Tamagotchi) face() string {
	switch {
	case t.health == 0:
		return "😵" // dead face
	case t.health < 30:
		return "😷" // sick face
	case t.happiness < 30:
		return "😢" // sad face
	case t.hunger > 70:
		return "😫" // hungry face
	case t.sleepy > 70:
		return "😴" // sleepy face
	default:
		return "😃" // happy face
	}
}

// Markdown method for Tamagotchi
func (t *Tamagotchi) Markdown() string {
	t.update()
	return ufmt.Sprintf(`# %s %s

* age: %d
* hunger: %d
* happiness: %d
* health: %d
* sleepy: %d`,
		t.name, t.Face(),
		t.age, t.hunger, t.happiness, t.health, t.sleepy,
	)
}

func bound(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
