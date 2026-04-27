package bt

// Blackboard holds the user-defined state T passed into every tick.
// Pass as a pointer so leaf functions can mutate Data in place.
type Blackboard[T any] struct {
	Data T
}
