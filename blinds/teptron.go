package blinds

import (
	"context"

	"github.com/tombooth/home"
)

type MoveBlind struct {
	DefaultBlind

	csrMesh *CSRMesh

	name  string
	state home.State
}

func setPositionCommand(position uint8) ([]byte, error) {
	return toBytes([]interface{}{
		uint8(0),    // Object ID
		uint8(0),    // Flag
		uint8(0x73), // Magic
		uint8(0x22), // Move command
		position,
	})
}

func NewMoveBlind(name, destination string, pin int, state home.State) Blind {
	return &MoveBlind{
		csrMesh: NewCSRMesh(destination, pin),
		name:    name,
		state:   state,
	}
}

func (c *MoveBlind) Name() string {
	return c.name
}

func (c *MoveBlind) Id() home.DeviceId {
	return home.DeviceId(c.csrMesh.destination)
}

func (c *MoveBlind) State(ctx context.Context) (home.State, error) {
	return c.state, nil
}

/*func (c *MoveBlind) Thing() error {
    if command, err := toBytes([]interface{}{
		uint8(0),    // Object ID
		uint8(0),    // Flag
		uint8(0x73), // Magic
        uint8(0x10), // Command
	}); err != nil {
        return err
    } else if err := c.csrMesh.Send(0x0021, command); err != nil {
        return err
    } else {
        return nil
    }
}*/

func (c *MoveBlind) Open(ctx context.Context) error {
	if command, err := setPositionCommand(0); err != nil {
		return err
	} else if err := c.csrMesh.Send(ctx, 0x0021, command); err != nil {
		return err
	} else {
		c.state = NewBlindState(true)
		return nil
	}
}

func (c *MoveBlind) Close(ctx context.Context) error {
	if command, err := setPositionCommand(255); err != nil {
		return err
	} else if err := c.csrMesh.Send(ctx, 0x0021, command); err != nil {
		return err
	} else {
		c.state = NewBlindState(false)
		return nil
	}
}
