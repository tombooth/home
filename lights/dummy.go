package lights

import (
	"context"
	"time"

	"github.com/tombooth/home"
)

type dummyLight struct {
	DefaultLight

	id    home.DeviceId
	name  string
	state home.State
}

func NewDummyLight(id home.DeviceId, name string, isOn bool) Light {
	return &dummyLight{
		id:    id,
		name:  name,
		state: NewLightState(isOn),
	}
}

func (l *dummyLight) Id() home.DeviceId { return l.id }
func (l *dummyLight) Name() string      { return l.name }

func (l *dummyLight) State(ctx context.Context) (home.State, error) {
	return l.state, nil
}

func (l *dummyLight) TurnOn(ctx context.Context) error {
	l.state = NewLightState(true)
	return nil
}

func (l *dummyLight) TurnOff(ctx context.Context) error {
	l.state = NewLightState(false)
	return nil
}

type slowLight struct {
	dummyLight
}

func NewSlowLight(id home.DeviceId, name string, isOn bool) Light {
	return &slowLight{
		dummyLight: dummyLight{
			id:    id,
			name:  name,
			state: NewLightState(isOn),
		},
	}
}

func (l *slowLight) State(ctx context.Context) (home.State, error) {
	delay := time.NewTimer(5 * time.Second)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-delay.C:
		return l.state, nil
	}
}
