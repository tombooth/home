package blinds

import (
	"bytes"
	"encoding/binary"
)

func toBytes(data []interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	for _, v := range data {
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			return []byte{}, err
		}
	}

	return buf.Bytes(), nil
}
