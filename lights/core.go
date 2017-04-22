package lights

import (
	"encoding/json"

    "github.com/tombooth/home"
)




const LightType = "light"

type Light interface {
    home.Device

    TurnOn() error
    TurnOff() error
}

type DefaultLight struct { }

func (d *DefaultLight) Type() home.DeviceType {
    return LightType
}

func (d *DefaultLight) UnmarshalState(raw []byte) (home.State, error) {
	var state lightState

	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, err
	}

	return &state, nil
}





type LightState interface {
    home.State

    On() bool
}

type lightState struct {
    IsOn bool `json:"on"`
}

func NewLightState(on bool) home.State {
    return &lightState{on}
}

func (ls *lightState) On() bool {
    return ls.IsOn
}





type LightController struct { }

func (_ *LightController) For() home.DeviceType {
    return LightType
}

func (_ *LightController) Reconcile(l home.Device, c, d home.State) error {
    light, current, desired := l.(Light), c.(LightState), d.(LightState)

    if current.On() != desired.On() {
        if desired.On() {
            return light.TurnOn()
        } else {
            return light.TurnOff()
        }
    }

    return nil
}
