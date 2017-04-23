package blinds

import (
	"context"
	"encoding/json"

    "github.com/tombooth/home"
)




const BlindType = "blind"

type Blind interface {
    home.Device

    Open(context.Context) error
    Close(context.Context) error
}

type DefaultBlind struct { }

func (d *DefaultBlind) Type() home.DeviceType {
    return BlindType
}

func (d *DefaultBlind) UnmarshalState(raw []byte) (home.State, error) {
	var state blindState

	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, err
	}

	return &state, nil
}





type BlindState interface {
    home.State

    Open() bool
}

type blindState struct {
    IsOpen bool `json:"open"`
}

func NewBlindState(open bool) home.State {
    return &blindState{open}
}

func (bs *blindState) Open() bool {
    return bs.IsOpen
}





type BlindController struct { }

func (_ *BlindController) For() home.DeviceType {
    return BlindType
}

func (_ *BlindController) Reconcile(ctx context.Context, b home.Device, c, d home.State) error {
    blind, current, desired := b.(Blind), c.(BlindState), d.(BlindState)

    if current.Open() != desired.Open() {
        if desired.Open() {
            return blind.Open(ctx)
        } else {
            return blind.Close(ctx)
        }
    }

    return nil
}
