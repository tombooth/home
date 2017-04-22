package lights

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "regexp"
    "strings"

    "github.com/tombooth/home"

    "github.com/tidwall/gjson"
)



func findLine(output, pattern string) string {
    lines := strings.Split(output, "\n")
    for _, line := range lines {
        matched, _ := regexp.MatchString(pattern, line)
        if matched {
            return line
        }
    }
    return ""
}




type IkeaGateway struct {
    uri string
    key string
}

func NewIkeaGateway(gatewayAddress, key string) *IkeaGateway {
    return &IkeaGateway{
        uri: fmt.Sprintf("coaps://%s:5684", gatewayAddress),
        key: key,
    }
}

func (g *IkeaGateway) coap(method, path, payload string) (string, error) {
    cmd := exec.Command(
        "/usr/local/bin/coap-client",
        "-m", method,
        "-u", "Client_identity",
        "-k", g.key,
        "-e", payload,
        fmt.Sprintf("%s%s", g.uri, path),
    )

    output, err := cmd.Output()
    if err != nil {
        return "", nil
    }

    return findLine(string(output), `^[\[\{].+$`), nil
}

func (g *IkeaGateway) AddressLight(id int) (Light, error) {
	path := fmt.Sprintf("/15001/%d", id)

    if deviceJson, err := g.coap("get", path, ""); err != nil {
		return nil, err
	} else {
		name := gjson.Get(deviceJson, "9001").String()
		return NewIkeaLight(g, id, name), nil
	}
}

func (g *IkeaGateway) AddressGroup(id int) (home.Group, error) {
	path := fmt.Sprintf("/15004/%d", id)

    if groupJson, err := g.coap("get", path, ""); err != nil {
		return home.Group{}, err
	} else {
		name := gjson.Get(groupJson, "9001").String()
		lights := []home.DeviceReference{}

		for _, light := range gjson.Get(groupJson, "9018.15002.9003").Array() {
			lights = append(lights, home.DeviceReference{
				LightType,
				home.DeviceId(fmt.Sprintf("%d", light.Int())),
			})
		}

		return home.Group{home.GroupId(fmt.Sprintf("%d", id)), LightType, name, lights}, nil
	}
}

func (g *IkeaGateway) AllLights() ([]home.Device, error) {
    coapJson, err := g.coap("get", "/15001", "")
    if err != nil {
        return []home.Device{}, err
    }

    var ids []int
    json.Unmarshal([]byte(coapJson), &ids)

    lights := []home.Device{}

    for _, id := range ids {
		if light , err := g.AddressLight(id); err != nil {
			return []home.Device{}, err
		} else {
			lights = append(lights, light)
		}
    }

    return lights, nil
}

func (g *IkeaGateway) AllGroups() ([]home.Group, error) {
    coapJson, err := g.coap("get", "/15004", "")
    if err != nil {
        return []home.Group{}, err
    }

    var ids []int
    json.Unmarshal([]byte(coapJson), &ids)

    groups := []home.Group{}

    for _, id := range ids {
    	if group , err := g.AddressGroup(id); err != nil {
			return []home.Group{}, err
		} else {
			groups = append(groups, group)
		}
    }

    return groups, nil
}







type IkeaLight struct {
    DefaultLight

    gateway *IkeaGateway
    path string

    id home.DeviceId
    name string
}

func NewIkeaLight(gateway *IkeaGateway, id int, name string) *IkeaLight {
    return &IkeaLight{
        gateway: gateway,
        path: fmt.Sprintf("/15001/%d", id),
        id: home.DeviceId(fmt.Sprintf("%d", id)),
		name: name,
    }
}


func (l *IkeaLight) State() (home.State, error) {
    if deviceJson, err := l.gateway.coap("get", l.path, ""); err != nil {
        return nil, err
    } else {
        return &lightState{
            gjson.Get(deviceJson, "3311.0.5850").Int() == 1,
        }, nil
    }
}

func (l *IkeaLight) TurnOn() error {
    _, err := l.gateway.coap("put", l.path, `{ "3311": [{ "5850": 1 }] }`)
    return err
}

func (l *IkeaLight) TurnOff() error {
    _, err := l.gateway.coap("put", l.path, `{ "3311": [{ "5850": 0 }] }`)
    return err
}

func (l *IkeaLight) Name() string {
    return l.name
}

func (l *IkeaLight) Id() home.DeviceId {
    return home.DeviceId(l.id)
}
