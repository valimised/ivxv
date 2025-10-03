package persistence

import (
	"fmt"

	"github.com/prologic/bitcask"
	"tivi.io/core/common/files"
	"tivi.io/core/voterlist"
)

const (
	MaxKeySize      = 32
	MaxDatafileSize = 512 * 1024 * 1024
)

type BitcaskVoterList struct {
	db   *bitcask.Bitcask
	path string
}

func NewBitcaskVoterList(path string) (*BitcaskVoterList, error) {

	if files.IsDir(path) != nil {
		return nil, fmt.Errorf("voter list at %s does not exist", path)
	}

	// actual db must be opened by a explicit call to Open
	return &BitcaskVoterList{
		db:   nil,
		path: path,
	}, nil
}

func (v *BitcaskVoterList) Open() (err error) {

	if v.db == nil {
		opts := []bitcask.Option{
			bitcask.WithMaxKeySize(MaxKeySize),
			bitcask.WithMaxDatafileSize(MaxDatafileSize),
		}

		db, err := bitcask.Open(v.path, opts...)

		if err != nil {
			return fmt.Errorf("OpenDB %s %v", v.path, err)
		}

		v.db = db
	}

	return
}

func (v *BitcaskVoterList) Path() string {
	return v.path
}

// Get return the value mapped to key in the persistent value. ok is false if
// no such key was found.
func (v *BitcaskVoterList) Voter(key string) (voter *voterlist.Voter, err error) {

	if v.db == nil {
		return nil, fmt.Errorf("Voterlist DB is not open")
	}

	var rawkey []byte
	var rawvalue []byte

	rawkey, err = voterlist.GetRawKey(key)
	if err != nil {
		return
	}

	if !v.db.Has(rawkey) {
		return nil, nil
	}

	rawvalue, err = v.db.Get(rawkey)
	if err != nil {
		return
	}

	return voterlist.NewVoterFromKeyValue(rawkey, rawvalue)
}

func (v *BitcaskVoterList) Count() (uint, error) {

	if v.db == nil {
		return 0, fmt.Errorf("Voterlist DB is not open")
	}

	return uint(v.db.Len()), nil
}

func (v *BitcaskVoterList) Close() {
	if v.db != nil {
		v.db.Close()
		v.db = nil
	}
}
