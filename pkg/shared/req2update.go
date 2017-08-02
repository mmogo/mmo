package shared

import "fmt"

// ToUpdate converts requests (specifically their non-nil field) to updates
// ToUpdate does not do request validation,
// Assumes the request is valid
// If request is not a known type, ToUpdate will panic
func ToUpdate(sourceID string, reqContent interface{}) *Update {
	switch content := reqContent.(type) {
	case *MoveRequest:
		return &Update{PlayerDestination: &PlayerDestination{
			ID:          sourceID,
			Destination: content.Destination,
		}}
	case *SpeakRequest:
		return &Update{PlayerSpoke: &PlayerSpoke{
			ID:   sourceID,
			Text: content.Text,
		}}
	}
	panic(fmt.Sprintf("unknown request type %#v", reqContent))
}
