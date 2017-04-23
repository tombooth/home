package blinds

import (
	"context"
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os/exec"
	//"github.com/currantlabs/ble"
	//"github.com/currantlabs/ble/examples/lib/dev"
)

type CSRMesh struct {
	destination string
	key         []byte
}

func NewCSRMesh(destination string, pin int) *CSRMesh {
	return &CSRMesh{
		destination: destination,
		key:         pinToKey(pin),
	}
}

func freshSeq() uint32 {
	out := make([]byte, 1)
	rand.Read(out)
	return uint32(out[0])
}

func (csr *CSRMesh) Send(ctx context.Context, handle uint16, data []byte) error {
	packet, err := makePacket(
		freshSeq(),
		csr.key,
		data,
	)

	if err != nil {
		return err
	}

	if len(packet) > 20 {
		if err = csr.write(ctx, handle, packet[0:20]); err != nil {
			return err
		}

		if err = csr.write(ctx, handle+3, packet[20:]); err != nil {
			return err
		}
	} else {
		if err = csr.write(ctx, handle, packet); err != nil {
			return err
		}
	}

	return nil
}

/*func findByHandle(client ble.Client, handle uint16) (*ble.Characteristic, error) {
	services, err := client.DiscoverServices(nil)
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		chars, err := client.DiscoverCharacteristics(nil, service)
		if err != nil {
			return nil, err
		}

		for _, char := range chars {
			if char.ValueHandle == handle {
				return char, nil
			}
		}
	}

	return nil, fmt.Errorf("Couldn't find characteristic with handle '%d'", handle)
}

func (csr *CSRMesh) write(handle uint16, data []byte) error {

	d, err := dev.NewDevice("default")
	if err != nil {
		return err
	}
	ble.SetDefaultDevice(d)

	//addr := ble.NewAddr("E4288BA3-C43B-42C6-8359-6E5886292CBC") // need to switch for MAC
	addr := ble.NewAddr(csr.destination)

	client, err := ble.Dial(context.Background(), addr)
	if err != nil {
		return err
	}

	char, err := findByHandle(client, handle)
	if err != nil {
		return err
	}

	err = client.WriteCharacteristic(char, data, true)
	if err != nil {
		return err
	}

	return nil
}*/

func (csr *CSRMesh) write(ctx context.Context, handle uint16, data []byte) error {
	handleBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(handleBytes, handle)
	handleHex := make([]byte, hex.EncodedLen(2))
	hex.Encode(handleHex, handleBytes)

	dataHex := make([]byte, hex.EncodedLen(len(data)))
	hex.Encode(dataHex, data)

	errOut := make(chan error)

	go func() {
		cmd := exec.Command(
			"/usr/bin/gatttool",
			"-b", csr.destination,
			"--char-write-req",
			"-a", fmt.Sprintf("0x%s", handleHex),
			"-n", string(dataHex),
		)

		fmt.Printf("/usr/bin/gatttool -b %s --char-write-req -a 0x%s -n %s\n", csr.destination, handleHex, dataHex)

		output, err := cmd.Output()

		fmt.Println(string(output))

		errOut <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errOut:
		return err
	}
}

func reverse(numbers []byte) []byte {
	newNumbers := make([]byte, len(numbers))
	for i, j := 0, len(numbers)-1; i < j; i, j = i+1, j-1 {
		newNumbers[i], newNumbers[j] = numbers[j], numbers[i]
	}
	return newNumbers
}

func pinToKey(pin int) []byte {
	strPin := fmt.Sprintf("%04d", pin)
	bytesPin := append([]byte(strPin), byte(0))
	bytesPin = append(bytesPin, []byte("MCP")...)

	h := sha256.New()
	h.Write(bytesPin)
	out := reverse(h.Sum(nil))
	return out[:16]
}

func xor(dst, a, b []byte) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
	return n
}

func makePacket(seq uint32, key, data []byte) ([]byte, error) {
	base, err := toBytes([]interface{}{
		seq,
		uint8(0),
		uint8(0x80), // Magic
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
	})

	if err != nil {
		return []byte{}, err
	}

	if len(key) != 16 || len(base) != 16 {
		return []byte{}, fmt.Errorf("We are assuming byte lengths of 16, because fake ECB and stuff")
	}

	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	encryptedBase := make([]byte, 16)
	aesCipher.Encrypt(encryptedBase, base)

	payload := make([]byte, len(data))
	xor(payload, data, encryptedBase[:len(data)])

	message, err := toBytes([]interface{}{
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		uint8(0),
		seq,
		uint8(0x80),
		payload,
	})

	if err != nil {
		return []byte{}, err
	}

	sigHash := hmac.New(sha256.New, key)
	sigHash.Write(message)
	sig := reverse(sigHash.Sum(nil))[:8]

	return toBytes([]interface{}{
		seq,
		uint8(0x80),
		payload,
		sig,
		uint8(0xFF),
	})
}
