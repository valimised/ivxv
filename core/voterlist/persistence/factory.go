package persistence

import (
	"fmt"

	"tivi.io/core/voterlist"
)

func GetVoterListImpl(name string, path string) (vl voterlist.VoterList, err error) {

	switch name {
	case "bolt":
		vl, err = NewBoltVoterList(path)
	case "bitcask":
		vl, err = NewBitcaskVoterList(path)
	default:
		err = fmt.Errorf(
			"provided voterlist backend option is not supported: %s", name)
	}

	return
}
