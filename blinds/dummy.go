package blinds

import (
	"context"

	"github.com/tombooth/home"
)

type dummyBlind struct {
	DefaultBlind

	id    home.DeviceId
	name  string
	state home.State
}

func NewDummyBlind(id home.DeviceId, name string, isOn bool) Blind {
	return &dummyBlind{
		id:    id,
		name:  name,
		state: NewBlindState(isOn),
	}
}

func (l *dummyBlind) Id() home.DeviceId { return l.id }
func (l *dummyBlind) Name() string      { return l.name }
func (l *dummyBlind) State(_ context.Context) (home.State, error) {
	return l.state, nil
}

func (l *dummyBlind) Open(_ context.Context) error {
	l.state = NewBlindState(true)
	return nil
}

func (l *dummyBlind) Close(_ context.Context) error {
	l.state = NewBlindState(false)
	return nil
}
