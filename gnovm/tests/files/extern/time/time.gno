package time

type Time struct {
	wall uint64
	ext  int64
	loc  *Location
}

// XXX dummy
func (t Time) Minute() int {
	return int(1111)
}

// XXX dummy
func (t Time) Second() int {
	return int(2222)
}

// XXX dummy
func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time {
	return Time{
		wall: 0,
		ext:  0,
		loc:  nil,
	}
}

var (
	UTC    *Location = &utcLoc
	utcLoc           = Location{name: "UTC"}
)

type Location struct {
	name       string
	zone       []zone
	tx         []zoneTrans
	extend     string
	cacheStart int64
	cacheEnd   int64
	cacheZone  *zone
}

type zone struct {
	name   string
	offset int
	isDST  bool
}

type zoneTrans struct {
	when         int64
	index        uint8
	isstd, isutc bool
}

type Month int

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)
