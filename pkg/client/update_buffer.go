package client

import (
	"time"

	"github.com/mmogo/mmo/pkg/shared"
)

// updatebuffer is a time-sorted slice of updates
type UpdateBuffer []*shared.Update

func (b UpdateBuffer) Contains(update *shared.Update) bool {
	for _, up := range b {
		if up == update {
			return true
		}
	}
	return false
}

func (b UpdateBuffer) Insert(update *shared.Update) UpdateBuffer {
	for i, up := range b {
		if up.Processed.Before(update.Processed) {
			return append(append(b[:i], update), b[i:]...)
		}
	}
	return append(b, update)
}

func (b UpdateBuffer) From(t time.Time) UpdateBuffer {
	for i, update := range b {
		if !update.Processed.Before(t) {
			return b[i:]
		}
	}
	return UpdateBuffer{}
}
