package web

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tombooth/home"

	"github.com/julienschmidt/httprouter"
)

func response(w http.ResponseWriter, data interface{}) {
	out, _ := json.Marshal(data)
	fmt.Fprint(w, string(out))
}

type StateResource struct {
	Id home.DeviceId
	Type home.DeviceType
	Name string
	State home.State
}

func deviceFromMap(dmap map[home.DeviceId]home.Device) home.Device {
	for _, device := range dmap {
		return device
	}
	return nil
}

func Routes(world *home.World, groups []home.Group, transitions chan home.WorldTransition) http.Handler {
	router := httprouter.New()

	router.GET("/device", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		types := []home.DeviceType{}

		for dtype, _ := range world.Desired {
			types = append(types, dtype)
		}

		response(w, types)
	})
	router.GET("/device/:type", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		states := []StateResource{}
		dtype := home.DeviceType(params.ByName("type"))

		for id, state := range world.Desired[dtype] {
			device := world.Devices[dtype][id]
			states = append(states, StateResource{id, dtype, device.Name(), state})
		}

		response(w, states)
	})
	router.POST("/device/:type", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		dtype := home.DeviceType(params.ByName("type"))
		devicesMap := world.Devices[dtype]
		device := deviceFromMap(devicesMap)

		rawState, _ := ioutil.ReadAll(r.Body)
		newState, _ := device.UnmarshalState(rawState)

		for id, _ := range devicesMap {
			transitions <- home.WorldTransition{home.DeviceReference{dtype, id}, newState}
		}

		response(w, "ok")
	})
	router.GET("/device/:type/:id", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		dtype := home.DeviceType(params.ByName("type"))
		id := home.DeviceId(params.ByName("id"))
		device := world.Devices[dtype][id]
		response(w, StateResource{id, dtype, device.Name(), world.Desired[dtype][id]})
	})
	router.POST("/device/:type/:id", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		dtype := home.DeviceType(params.ByName("type"))
		id := home.DeviceId(params.ByName("id"))
		device := world.Devices[dtype][id]

		rawState, _ := ioutil.ReadAll(r.Body)
		newState, _ := device.UnmarshalState(rawState)

		transitions <- home.WorldTransition{home.DeviceReference{dtype, id}, newState}

		response(w, "ok")
	})

	groupsById := map[home.GroupId]home.Group{}

	for _, group := range groups {
		groupsById[group.Id] = group
	}

	router.GET("/group", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		response(w, groups)
	})
	router.GET("/group/:id", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		id := home.GroupId(params.ByName("id"))
		response(w, groupsById[id])
	})
	router.POST("/group/:id", func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		group := groupsById[home.GroupId(params.ByName("id"))]
		device := world.Devices[group.References[0].Type][group.References[0].Id]

		rawState, _ := ioutil.ReadAll(r.Body)
		newState, _ := device.UnmarshalState(rawState)

		for _, ref := range group.References {
			transitions <- home.WorldTransition{ref, newState}
		}

		response(w, "ok")
	})

	return router
}
