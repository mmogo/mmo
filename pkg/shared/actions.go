package shared

// Action describes the activity of an entity
type Action int

const (
	A_IDLE Action = iota
	A_WALK
	A_CHAT_
	A_SLASH
	A_SHOOT
	A_SPELL
	A_THRUST
	A_HURT
	A_DEAD
)
