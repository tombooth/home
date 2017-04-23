package home

import (
	"context"
)

type State interface{}

type DeviceType string

type DeviceId string

type DeviceReference struct {
	Type DeviceType
	Id   DeviceId
}

type Device interface {
	Id() DeviceId
	Type() DeviceType
	Name() string
	State(context.Context) (State, error)
	UnmarshalState([]byte) (State, error)
}

type GroupId string

type Group struct {
	Id         GroupId
	Type       DeviceType
	Name       string
	References []DeviceReference
}

type Controller interface {
	For() DeviceType
	Reconcile(context.Context, Device, State, State) error
}
