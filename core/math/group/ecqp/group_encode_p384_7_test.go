package ecqp

import (
	"reflect"
	"testing"
	"tivi.io/core/math/group/internal"
)

func TestGroupEncodeP384FieldEncodingBits7Positive(t *testing.T) {
	t.Log("Positive cases for encoding NIST-P384 group with field encoding bits 7")

	for _, message := range ecqPGroupP384MessagePositive {
		padded, err := ecqPGroupP384.PadBytes(message)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ecqPGroupP384.Encode(padded)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGroupEncodeP384FieldEncodingBits7Negative(t *testing.T) {
	t.Log("Negative cases for encoding NIST-P384 group with field encoding bits 7")

	for _, message := range ecqPGroupP384MessagePositive {
		padded, err := ecqPGroupP384.PadBytes(message)
		if err != nil {
			t.Fatal(err)
		}

		// Make padded message 2 bits (376+2=378) bigger so it won't fit it max allowed 377 (384-7=377)
		padded2 := make([]byte, 1)
		padded2[0] = 0x02
		padded2 = append(padded2, padded...)

		_, err = ecqPGroupP384.Encode(padded2)
		if err == nil {
			t.Fatal(err)
		}
		if reflect.TypeOf(err) != reflect.TypeOf(internal.MessageBitLenTooLarge{}) {
			t.Fatalf("expected error %v, but got %v", internal.MessageBitLenTooLarge{}, reflect.TypeOf(err))
		}
	}
}
