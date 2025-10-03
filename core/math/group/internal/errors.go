package internal

import "fmt"

const (
	ErrDifferentGenOrder    = "different generator orders"
	ErrDifferentFieldOrder  = "different field orders"
	ErrNotASN1OctetString   = "element is not an ASN.1 OCTET STRING"
	ErrTrailingBytesASN1    = "incomplete ASN.1 element"
	ErrScalarModCmpGenOrder = "scalar modulo != generator order"
	ErrLargeMessageBytes    = "message byte length is larger than allowed"
)

type UnexpectedPaddingByte struct {
	Prefix string
	Index  int
	Byte   byte
}

func (e UnexpectedPaddingByte) Error() string {
	return fmt.Sprintf("%s: unexpected padding byte %x at slice[%d]", e.Prefix, e.Byte, e.Index)
}

type UnexpectedPaddingHeader struct {
	Prefix string
	Byte   byte
}

func (e UnexpectedPaddingHeader) Error() string {
	return fmt.Sprintf("%s: unexpected padding header %x", e.Prefix, e.Byte)
}

type NoEncodedMessage struct {
	Prefix string
}

func (e NoEncodedMessage) Error() string {
	return fmt.Sprintf("%s: message is padded to the full length, no space for data left", e.Prefix)
}

type MessageBitLenTooLarge struct {
	Prefix       string
	MsgBitLen    int
	MaxMsgBitLen int
}

func (e MessageBitLenTooLarge) Error() string {
	return fmt.Sprintf("%s: message bit length %d is too large for maximum %d allowed", e.Prefix, e.MsgBitLen, e.MaxMsgBitLen)
}

type MessageByteLenTooLarge struct {
	Prefix        string
	MsgByteLen    int
	MaxMsgByteLen int
}

func (e MessageByteLenTooLarge) Error() string {
	return fmt.Sprintf("%s: message byte length %d is too large for maximum %d allowed", e.Prefix, e.MsgByteLen, e.MaxMsgByteLen)
}

type NotASN1OctetString struct {
	Prefix string
}

func (e NotASN1OctetString) Error() string {
	return fmt.Sprintf("%s: element is not an ASN.1 OCTET STRING", e.Prefix)
}

type NotASN1Integer struct {
	Prefix string
}

func (e NotASN1Integer) Error() string {
	return fmt.Sprintf("%s: element is not an ASN.1 INTEGER", e.Prefix)
}

type ASN1TrailingBytes struct {
	Prefix string
}

func (e ASN1TrailingBytes) Error() string {
	return fmt.Sprintf("%s: trailing bytes while parsing ASN.1 element", e.Prefix)
}

type DifferentOrder struct {
	Prefix string
}

func (e DifferentOrder) Error() string {
	return fmt.Sprintf("%s: groups have different generator order", e.Prefix)
}

type DifferentFieldOrder struct {
	Prefix string
}

func (e DifferentFieldOrder) Error() string {
	return fmt.Sprintf("%s: groups have different field order", e.Prefix)
}

type NotOnCurve struct {
	Prefix string
}

func (e NotOnCurve) Error() string {
	return fmt.Sprintf("%s: point is not on a curve", e.Prefix)
}

type ScalarModCmpOrder struct {
	Prefix string
}

func (e ScalarModCmpOrder) Error() string {
	return fmt.Sprintf("%s: scalar modulo not equal to generator order", e.Prefix)
}

type PaddedByteLen struct {
	Prefix          string
	MsgByteLen      int
	ExpectedByteLen int
}

func (e PaddedByteLen) Error() string {
	return fmt.Sprintf("%s: padded byte length %d is not equal to expected %d", e.Prefix, e.MsgByteLen, e.ExpectedByteLen)
}

type PaddedBitLen struct {
	Prefix         string
	MsgBitLen      int
	ExpectedBitLen int
}

func (e PaddedBitLen) Error() string {
	return fmt.Sprintf("%s: padded bit length %d is not equal to expected %d", e.Prefix, e.MsgBitLen, e.ExpectedBitLen)
}

type ElementLessThanOne struct {
	Prefix string
}

func (e ElementLessThanOne) Error() string {
	return fmt.Sprintf("%s: element value is less than 1", e.Prefix)
}
