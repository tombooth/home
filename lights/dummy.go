package lights

import (
	"github.com/tombooth/home"
)

type dummyLight struct {
	DefaultLight

	id home.DeviceId
	name string
	state home.State
}

func NewDummyLight(id home.DeviceId, name string, isOn bool) Light {
	return &dummyLight {
		id: id,
		name: name,
		state: NewLightState(isOn),
	}
}

func (l *dummyLight) Id() home.DeviceId { return l.id }
func (l *dummyLight) Name() string { return l.name }
func (l *dummyLight) State() (home.State, error) { return l.state, nil }

func (l *dummyLight) TurnOn() error {
	l.state = NewLightState(true)
	return nil
}

func (l *dummyLight) TurnOff() error {
	l.state = NewLightState(false)
	return nil
}
