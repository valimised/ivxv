package group

import (
	"fmt"
	"math/big"
)

// toBytes converts *big.Int msg to raw bytes, preserving all leading zero bits.
func toBytes(msg *big.Int) (raw []byte, err error) {
	// Empty message == empty byte slice
	if msg.BitLen() == 0 {
		raw = []byte{}
	} else {
		raw = msg.Bytes()
	}

	// From bits to bytes
	msgByteLen := (msg.BitLen() + 7) / 8
	result := make([]byte, msgByteLen)

	if len(raw) > 0 {
		copy(result[msgByteLen-len(raw):], raw)
	}

	return result, nil
}

// PadBits pads msg up until (group field bitlen - 2 - 1).
//
// Padded msg looks like:
//
//	0b0111... 0b110XX...XXX
//
// where first 0 and 1 bits are padding header, following by padding 1 bits,
// and after padding 1 bits there is a 0 padding end bit. X bits
// denote msg bits which are filled up until the end.
//
// NB! Currently only supported by ModqP and ECqP groups.
//
// The main idea behind PadBits is that message is padded by bits
// which therefore leaves no empty space in padded message as it would
// be with traditional byte padding, where message is byte-aligned.
func PadBits(msg []byte, fieldBitLen uint64) (padded []byte, err error) {
	// -2 to reserve padding header bits
	// -1 to reserve padding end bit
	msgMaxBitLen := fieldBitLen - 2 - 1

	msgInt := new(big.Int).SetBytes(msg)

	// 3 bits are reserved, while other bits can be occupied by msg
	if uint64(msgInt.BitLen()) > msgMaxBitLen {
		return nil, fmt.Errorf("math/group: msg bit length %v is larger than allowed %v", msgInt.BitLen(), msgMaxBitLen)
	}

	// Padding (header and end bits included) bits occupy the rest space that msg bits doesn't
	paddingBitLen := msgMaxBitLen - uint64(msgInt.BitLen())

	paddedInt := big.NewInt(1)

	// +2 to allocate padding header 0 and 1 bits, but
	// currently it looks like 0b1000..., so you can
	// see that first padding header bit is not 0
	paddedInt.Lsh(paddedInt, uint(paddingBitLen)+2)

	// This clever trick does 3 things:
	// a) sets padding header bits to 0b01
	// b) sets padding bits, i.e. 0b011111...
	// c) sets padding end bit to 0, i.e. 0b011111...0
	paddedInt.Sub(paddedInt, big.NewInt(2))

	// Allocates exactly msg bitlen bits
	paddedInt.Lsh(paddedInt, uint(msgInt.BitLen()))

	// Sets all msg bits to the allocated space
	paddedInt.Or(paddedInt, msgInt)

	// Returns byte slice with preserved leading zero bits
	return toBytes(paddedInt)
}

// UnpadBits strips bit padding from padded message.
//
// NB! Currently only supported by ModqP and ECqP groups.
func UnpadBits(padded []byte) (msg []byte, err error) {
	unpadded := new(big.Int).SetBytes(padded)

	// Most-significant (padding header 1 bit) bit should be 1, since padding header 0 bit is stripped by big.Int
	isMsbSet := unpadded.Bit(unpadded.BitLen() - 1)

	if isMsbSet != 1 {
		return nil, fmt.Errorf("math/group: padding header is not 0b01")
	}

	// Mask 0b0111...0, not to override existing bits in padding message
	mask := big.NewInt(1)
	mask.Lsh(mask, uint(unpadded.BitLen()))
	mask.Sub(mask, big.NewInt(1))

	// If we apply boolean NOT on a padded message, it will
	// mirror padding bits to be 0b1000...1, instead of initial 0b0111...0
	unpadded.Not(unpadded)

	// Now we apply 0b0111...0 mask, that will clear all header and padding bits,
	// only padding end bit will survive
	unpadded.And(unpadded, mask)

	// Now we check that padding end bit, inverted, is indeed 1
	isMsbSet = unpadded.Bit(unpadded.BitLen() - 1)

	if isMsbSet != 1 {
		return nil, fmt.Errorf("math/group: padding end bit is not 1")
	}

	// Now we create another mask to clear all 0 bits from a padded message,
	// which is currently has a form of 0b1XXX...X, where 1 is a padding end bit
	// and X are inverted message bits
	mask = big.NewInt(1)
	mask.Lsh(mask, uint(unpadded.BitLen()))
	mask.Sub(mask, big.NewInt(1))

	// Boolean NOT will invert padding end bit back to the original 0 value,
	// as well as message bits
	unpadded.Not(unpadded)
	unpadded.And(unpadded, mask)

	// Padding end bit is stripped by big.Int and only message bits are left behind
	return toBytes(unpadded)
}
