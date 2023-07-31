package adpattern

import "time"

type AdPattern struct {
	IsLTimeSet    bool
	IsRTimeSet    bool
	PublishedOnly bool
	AuthorID      int64
	LDate         time.Time
	RDate         time.Time
}
