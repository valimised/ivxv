package ecqp

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"tivi.io/core/math/group/internal"
)

var ecqPGroupP384MessagePositive = map[string][]byte{
	// Largest message is 368 bits, since PadBytes does byte-aligning on a message.
	// Actually maximum allowed message bits are 374 (384-2-1-7), so as you can see
	// 374-368=6 bits are left unused. That's why in a future we may switch to group.PadBits,
	// which doesn't align a message by bytes and instead uses raw bits.
	"largest": {0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},

	// Empty slice
	"empty slice": {},

	// Empty message
	"empty message": []byte(""),

	// Regular message 1
	"regular message 1": []byte("0000.101"),

	// Regular message 2
	"regular message 2": []byte("3011.678"),

	// Regular message 3
	"regular message 3": []byte("1111.2018"),

	// Regular message 4
	"regular message 4": []byte("Candidate\x1FParty\x1FLogo\x1FGünnar Metsäpuu"),

	// All 368 bits of zero bytes message
	"empty zero bytes": {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},

	// One zero byte message
	"empty zero byte": {0x00},
}

var ecqPGroupP384MessageNegative = map[string][]byte{
	// Message is exactly 374 bits (384-2-1-7), but since it is byte-aligned (376 bits),
	// maximum allowed is 368 bits
	"374bits": {0x3F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},

	// Message is exactly 373 bits (even smaller than (384-2-1-7)), but since it is
	// byte-aligned (376 bits), maximum allowed is 368 bits
	"373bits": {0x1F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},

	// Message is exactly 376 bits, which is greater than even maximum allowed theoretical
	// 384-2-1-7=374 bits
	"376bits": {0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},

	// 47 bytes of 0x00, while maximum allowed is 46 bytes (368/8)
	"47bytes": {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
}

func TestGroupPadBytesP384FieldEncodingBits7Positive(t *testing.T) {
	t.Log("Positive cases for padding NIST-P384 group with field encoding bits 7")

	// Field size is 384 bits, i.e. 48 bytes. Padding reserves 2 bits for padding head
	// and 1 bit for padding end. Also it reserves 7 bits for field encoding.
	// Which gives us total of 384-2-1-7=374 bits allowed for message encoding.
	// Since message is byte-aligned in PadBytes, there is no other option for
	// padding to be also byte-aligned. 374/8 we get 46.75 bytes. Padding should be
	// aligned to 47 bytes, i.e. 376 bits.
	resultByteLen := 47
	resultBitLen := 376

	for _, message := range ecqPGroupP384MessagePositive {
		padded, err := ecqPGroupP384.PadBytes(message)
		if err != nil {
			t.Fatal(err)
		}

		if len(padded) != resultByteLen {
			t.Fatalf("expected bytelen %d, got %d", resultByteLen, len(padded))
		}

		paddedInt := new(big.Int).SetBytes(padded)

		if paddedInt.BitLen() != resultBitLen {
			t.Fatalf("expected bitlen %d, got %d", resultBitLen, paddedInt.BitLen())
		}
	}
}

func TestGroupPadBytesP384FieldEncodingBits7Negative(t *testing.T) {
	t.Log("Negative cases for padding NIST-P384 group with field encoding bits 7")

	for _, message := range ecqPGroupP384MessageNegative {
		_, err := ecqPGroupP384.PadBytes(message)
		if err == nil {
			t.Fatal("expected error, but got nil")
		}
		if reflect.TypeOf(err) != reflect.TypeOf(internal.MessageBitLenTooLarge{}) {
			t.Fatalf("expected error %v, but got %v", internal.MessageBitLenTooLarge{}, reflect.TypeOf(err))
		}
	}
}

func TestGroupUnpadBytesP384FieldEncodingBits7Positive(t *testing.T) {
	t.Log("Positive cases for unpadding NIST-P384 group with field encoding bits 7")

	for _, message := range ecqPGroupP384MessagePositive {
		padded, err := ecqPGroupP384.PadBytes(message)
		if err != nil {
			t.Fatal(err)
		}
		unpadded, err := ecqPGroupP384.UnpadBytes(padded)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(message, unpadded) {
			t.Fatal("expected initial and unpadded messages to be equal")
		}
	}
}

func TestGroupUnpadBytesP384FieldEncodingBits7Negative(t *testing.T) {
	t.Log("Negative cases for unpadding NIST-P384 group with field encoding bits 7")

	for title, message := range ecqPGroupP384MessagePositive {
		padded, err := ecqPGroupP384.PadBytes(message)
		if err != nil {
			t.Fatal(err)
		}

		// Backup for later use
		backup := padded[0]

		// Set msb to 0b10111110, so when UnpadBytes checks header it will encounter
		// 0b01011111 (0xBE >> 1)
		padded[0] = 0xBE

		_, err = ecqPGroupP384.UnpadBytes(padded)
		if err == nil {
			t.Fatal("expected error, but got nil")
		}
		if reflect.TypeOf(err) != reflect.TypeOf(internal.UnexpectedPaddingHeader{}) {
			t.Fatalf("expected error %v, but got %v", internal.UnexpectedPaddingHeader{}, reflect.TypeOf(err))
		}

		padded[0] = backup

		if title != "largest" && title != "empty zero bytes" {
			// Set second byte to 0b01111010, so when UnpadBytes checks padding bytes it will report
			// that second byte is invalid
			backup = padded[1]
			padded[1] = 0x7A

			_, err = ecqPGroupP384.UnpadBytes(padded)
			if err == nil {
				t.Fatal("expected error, but got nil")
			}

			if reflect.TypeOf(err) != reflect.TypeOf(internal.UnexpectedPaddingByte{}) {
				t.Fatalf("expected error %v, but got %v", internal.UnexpectedPaddingByte{}, reflect.TypeOf(err))
			}

			padded[1] = backup
			for i := range padded[1:] {
				padded[i+1] = 0xFF
			}

			if _, err = ecqPGroupP384.UnpadBytes(padded); err == nil {
				t.Fatal("expected error, but got nil")
			}

			if reflect.TypeOf(err) != reflect.TypeOf(internal.NoEncodedMessage{}) {
				t.Fatalf("expected error %v, but got %v", internal.NoEncodedMessage{}, reflect.TypeOf(err))
			}
		}
	}
}
