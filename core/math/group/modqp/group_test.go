package modqp

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"
	"tivi.io/core/math/group"
	"tivi.io/core/math/group/internal"
)

func init() {
	name := "RFC3526ModPGroup3072"
	var err error
	modqPGroup3072, err = group.Get(name)
	if err != nil {
		panic(err)
	}
}

var modqPGroup3072 group.Group

var modqPGroup23 = &modqPGroup{name: "RFC3526ModPGroup23", groupOrder: big.NewInt(11), fieldOrder: big.NewInt(23), generator: big.NewInt(2)}

var positive3072Msgs = [][]byte{
	generateBytes(384-2-1, 0xFF),
	generateBytes(384-2-1, 0x00),
	{},
	[]byte(""),
	[]byte("0000.101"),
	[]byte("Candidate\x1FParty\x1FLogo\x1FGünnar Metsäpuu"),
	{0x00},
	[]byte("1111.2018"),
}

func generateBytes(size int, b byte) []byte {
	padding := make([]byte, size)
	// Fill the array with some values (e.g., index values)
	for i := range size {
		padding[i] = b
	}
	return padding
}

func TestGroupEqualDifferentOrder(t *testing.T) {
	t.Log("Equality checks should fail for different groups, but succeed for the same ones")

	var err error

	if err = modqPGroup3072.Equal(modqPGroup23); err == nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.DifferentOrder{}) {
		t.Fatalf("expected error %v, but got %v", internal.DifferentOrder{}, reflect.TypeOf(err))
	}
}

func TestGroupEqualDifferentFieldOrder(t *testing.T) {
	t.Log("Equality checks should fail for different groups, but succeed for the same ones")

	var err error

	modqPGroup23.groupOrder = modqPGroup3072.Order()

	if err = modqPGroup3072.Equal(modqPGroup23); err == nil {
		t.Fatal(err)
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.DifferentFieldOrder{}) {
		t.Fatalf("expected error %v, but got %v", internal.DifferentFieldOrder{}, reflect.TypeOf(err))
	}
}

func TestGroupEqualDifferentGenerator(t *testing.T) {
	t.Log("Equality checks should fail for different groups, but succeed for the same ones")

	var err error

	modqPGroup23.groupOrder = modqPGroup3072.Order()
	modqPGroup23.fieldOrder = modqPGroup3072.FieldOrder()
	modqPGroup23.generator = big.NewInt(1)
	if err = modqPGroup3072.Equal(modqPGroup23); err == nil {
		t.Fatal(err)
	}
}

func TestGroupPadBytes3072(t *testing.T) {
	for _, padded := range positive3072Msgs {
		padded2, err := modqPGroup3072.PadBytes(padded)
		if err != nil {
			t.Fatal(err)
		}

		if len(padded2) != 384 {
			t.Fatalf("expected %d bytelen, got %d", 384, len(padded2))
		}
	}

	padded := generateBytes(382, 0xFF)

	_, err := modqPGroup3072.PadBytes(padded)
	if err == nil {
		t.Fatal("expected error, got nil") // too large
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.MessageByteLenTooLarge{}) {
		t.Fatalf("expected error %v, but got %v", internal.MessageByteLenTooLarge{}, reflect.TypeOf(err))
	}
}

func TestGroupUnpadBytes3072(t *testing.T) {
	for _, msg := range positive3072Msgs {
		padded2, err := modqPGroup3072.PadBytes(msg)
		if err != nil {
			t.Fatal(err)
		}

		if len(padded2) != 384 {
			t.Fatalf("expected %d bytelen, got %d", 384, len(padded2))
		}
		unpadded, err := modqPGroup3072.UnpadBytes(padded2)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(unpadded, msg) {
			t.Fatal("initial message not equal to unpadded")
		}

	}
	// Grow by 1 byte
	padded2 := generateBytes(385, 0xFF)

	_, err := modqPGroup3072.UnpadBytes(padded2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.PaddedByteLen{}) {
		t.Fatalf("expected error %v, but got %v", internal.PaddedByteLen{}, reflect.TypeOf(err))
	}

	padded2 = padded2[:len(padded2)-1]
	padded2[0] = 0x01

	_, err = modqPGroup3072.UnpadBytes(padded2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	padded2[0] = 0x00
	padded2[1] = 0x00

	_, err = modqPGroup3072.UnpadBytes(padded2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	padded2[1] = 0x01
	padded2[2] = 0xFF
	padded2[24] = 0x24
	_, err = modqPGroup3072.UnpadBytes(padded2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.UnexpectedPaddingByte{}) {
		t.Fatalf("expected error %v, but got %v", internal.UnexpectedPaddingByte{}, reflect.TypeOf(err))
	}

	padded2[24] = 0xFF
	_, err = modqPGroup3072.UnpadBytes(padded2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if reflect.TypeOf(err) != reflect.TypeOf(internal.NoEncodedMessage{}) {
		t.Fatalf("expected error %v, but got %v", internal.NoEncodedMessage{}, reflect.TypeOf(err))
	}
}
