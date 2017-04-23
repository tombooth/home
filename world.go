package home

import (
	"context"
	"fmt"
)

type WorldState map[DeviceType]map[DeviceId]State

type WorldTransition struct {
	Reference DeviceReference
	NewState  State
}

type World struct {
	Controllers map[DeviceType]Controller
	Devices     map[DeviceType]map[DeviceId]Device

	Desired WorldState
}

func NewWorld() *World {
	return &World{
		Controllers: map[DeviceType]Controller{},
		Devices:     map[DeviceType]map[DeviceId]Device{},
	}
}

func (w *World) AddController(c Controller) error {
	if _, ok := w.Controllers[c.For()]; ok {
		return fmt.Errorf("Already have a controller for '%s'", c.For())
	}

	w.Controllers[c.For()] = c

	return nil
}

func (w *World) AddDevice(d Device) error {
	if _, ok := w.Devices[d.Type()]; !ok {
		w.Devices[d.Type()] = map[DeviceId]Device{}
	}

	if _, ok := w.Devices[d.Type()][d.Id()]; ok {
		return fmt.Errorf("Already have a '%s' with id '%s'", d.Type(), d.Id())
	}

	w.Devices[d.Type()][d.Id()] = d

	return nil
}

func (w *World) AddDevices(devices []Device) error {
	for _, device := range devices {
		if err := w.AddDevice(device); err != nil {
			return err
		}
	}
	return nil
}

func (w *World) Snapshot(ctx context.Context) (WorldState, error) {
	snapshot := WorldState{}

	for deviceType, devices := range w.Devices {
		if _, ok := snapshot[deviceType]; !ok {
			snapshot[deviceType] = map[DeviceId]State{}
		}

		for deviceId, device := range devices {
			if state, err := device.State(ctx); err != nil {
				return snapshot, err
			} else {

				snapshot[deviceType][deviceId] = state
			}
		}
	}

	return snapshot, nil
}

func (w *World) Seed(ctx context.Context) error {
	if snapshot, err := w.Snapshot(ctx); err != nil {
		return err
	} else {
		w.Desired = snapshot
	}

	return nil
}

func (w *World) Step(ctx context.Context) error {
	currentWorld, err := w.Snapshot(ctx)
	if err != nil {
		return err
	}

	for deviceType, deviceStates := range currentWorld {
		controller, ok := w.Controllers[deviceType]

		if !ok {
			return fmt.Errorf("No controller found for '%s'\n", deviceType)
		}

		for deviceId, state := range deviceStates {
			device, ok := w.Devices[deviceType][deviceId]
			if !ok {
				return fmt.Errorf("No device found for '%s/%s'", deviceType, deviceId)
			}

			desired, ok := w.Desired[deviceType][deviceId]
			if !ok {
				return fmt.Errorf("No desired state found for '%s/%s'", deviceType, deviceId)
			}

			if err := controller.Reconcile(ctx, device, state, desired); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *World) Apply(transition WorldTransition) error {
	w.Desired[transition.Reference.Type][transition.Reference.Id] = transition.NewState
	return nil
}
