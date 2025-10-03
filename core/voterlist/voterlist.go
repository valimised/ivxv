package voterlist

import (
	"encoding/hex"
	"fmt"
)

const (
	VoterNumberBinLength = 32
	VoterNumberHexLength = 64
)

var (
	ErrFormat = fmt.Errorf("Invalid voter format")
)

type VoterList interface {
	Open() error
	Path() string
	Voter(string) (*Voter, error)
	Count() (uint, error)
	Close()
}

type Voter struct {
	wkz   string
	hashA string
	hashB string
}

func GetRawKey(key string) (rawkey []byte, err error) {
	return hex.DecodeString(key)
}

func NewVoterFromString(wkz, n2, n3 string) (v *Voter, err error) {

	if len(wkz) != VoterNumberHexLength {
		err = ErrFormat
		return
	}
	if len(n2) != VoterNumberHexLength {
		err = ErrFormat
		return
	}
	if len(n3) != VoterNumberHexLength {
		err = ErrFormat
		return
	}

	v = &Voter{
		wkz:   wkz,
		hashA: n2,
		hashB: n3,
	}
	return
}

func NewVoterFromKeyValue(key, value []byte) (v *Voter, err error) {

	if len(key) != VoterNumberBinLength {
		err = ErrFormat
		return
	}

	if len(value) != 2*VoterNumberBinLength {
		err = ErrFormat
		return
	}

	wkz := hex.EncodeToString(key)
	n2 := hex.EncodeToString(value[:32])
	n3 := hex.EncodeToString(value[32:])

	v = &Voter{
		wkz:   wkz,
		hashA: n2,
		hashB: n3,
	}
	return
}

func (v *Voter) KeyValue() (key []byte, value []byte, err error) {

	if key, err = hex.DecodeString(v.wkz); err != nil {
		return
	}

	if value, err = hex.DecodeString(v.hashA + v.hashB); err != nil {
		return
	}

	return
}

func (v *Voter) GetDOBHash() string {
	return v.hashA
}

func (v *Voter) HashB() string {
	return v.hashB
}
