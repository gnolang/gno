package present

func init() {
	_ = Set(
		"demo",        // id
		"Demo Slides", // title
		"Demo Event",  // event
		"@demo",       // author
		"2025-02-02",  // date
		`## Slide One
- Point A
- Point B
- Point C

## Slide Two
- Feature 1
- Feature 2
- Feature 3

## Slide Three
- Next step
- Another step
- Final step`,
	)
}
