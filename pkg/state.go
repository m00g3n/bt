package bt

import "fmt"

// State represents the result of a behavior tree node tick.
type State int

const (
	Success State = 0
	Failure State = 1
)

func (s State) String() string {
	switch s {
	case Success:
		return "Success"
	case Failure:
		return "Failure"
	default:
		return fmt.Sprintf("State(%d)", int(s))
	}
}
