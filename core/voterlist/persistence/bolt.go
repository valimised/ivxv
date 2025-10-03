package persistence

import (
	"fmt"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"

	"tivi.io/core/common/files"
	"tivi.io/core/voterlist"
)

const (
	BucketName = "voters"
)

type BoltVoterList struct {
	db   *bbolt.DB
	path string
}

func NewBoltVoterList(path string) (*BoltVoterList, error) {

	dbpath := path

	if files.IsFile(dbpath) != nil {
		if files.IsDir(dbpath) == nil {
			dbpath = filepath.Join(dbpath, "dreinv-database.db")
			if files.IsFile(dbpath) != nil {
				return nil, fmt.Errorf("voter list at %s does not exist", dbpath)
			}
		} else {
			return nil, fmt.Errorf("voter list at %s does not exist", dbpath)
		}
	}

	return &BoltVoterList{
		db:   nil,
		path: dbpath,
	}, nil
}

func (v *BoltVoterList) Open() (err error) {

	if v.db == nil {
		db, err := bbolt.Open(v.path, 0600,
			&bbolt.Options{
				Timeout:  5 * time.Second,
				ReadOnly: true,
			})

		if err != nil {
			return fmt.Errorf("OpenDB %s %v", v.path, err)
		}

		v.db = db
	}

	return
}

func (v *BoltVoterList) Path() string {
	return v.path
}

func (v *BoltVoterList) Voter(key string) (voter *voterlist.Voter, err error) {

	if v.db == nil {
		return nil, fmt.Errorf("Voterlist DB is not open")
	}

	var rawkey []byte

	rawkey, err = voterlist.GetRawKey(key)
	if err != nil {
		return nil, err
	}

	v.db.View(func(tx *bbolt.Tx) error { // nolint:errcheck
		b := tx.Bucket([]byte(BucketName))
		rawvalue := b.Get(rawkey)
		if rawvalue == nil {
			return nil
		}
		voter, err = voterlist.NewVoterFromKeyValue(rawkey, rawvalue)
		return nil
	})

	return
}

func (v *BoltVoterList) Count() (count uint, err error) {

	if v.db == nil {
		return 0, fmt.Errorf("Voterlist DB is not open")
	}

	err = v.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(BucketName))
		count = uint(b.Stats().KeyN)
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (v *BoltVoterList) Close() {
	if v.db != nil {
		v.db.Close()
		v.db = nil
	}
}
