package localdb

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/dgraph-io/badger"
)

const (
	DefaultDBLocation string = "/badger/data"
)

var sequences = make(map[string]*badger.Sequence)

var localDB *LocalDB

type LocalDB struct {
	db *badger.DB
}

func GetHandle() (*LocalDB, error) {
	return GetHandleFromLocation(DefaultDBLocation)
}

func GetHandleFromLocation(dbLocation string) (*LocalDB, error) {
	myLock := sync.Mutex{}

	myLock.Lock()
	defer myLock.Unlock()

	if localDB != nil {
		return localDB, nil
	}

	info, err := os.Stat(dbLocation)
	if err != nil {
		if isNotPresent := os.IsNotExist(err); isNotPresent {
			return nil, fmt.Errorf("the directory %s need to exist", dbLocation)
		}
	}

	isMatch, _ := regexp.MatchString("drwx.*", info.Mode().String())
	if !isMatch {
		return nil, fmt.Errorf("the directory %s need to have read write execute permission", dbLocation)
	}

	badgerOpt := badger.DefaultOptions
	badgerOpt.Dir = dbLocation
	badgerOpt.ValueDir = dbLocation

	dbPtr, err := badger.Open(badgerOpt)
	if err != nil {
		return nil, fmt.Errorf("not able to initialise DB: %v", err)
	}
	localDB = &LocalDB{
		db: dbPtr,
	}

	return localDB, nil
}

func (l *LocalDB) Write(key string, data []byte) error {
	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

func (l *LocalDB) MultiWrite(multiData map[string][]byte) error {
	return l.db.Update(func(txn *badger.Txn) error {
		for key, val := range multiData {
			err := txn.Set([]byte(key), val)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (l *LocalDB) Read(key string) ([]byte, error) {
	var data = make([]byte, 0)
	err := l.db.View(func(txn *badger.Txn) error {
		item, readError := txn.Get([]byte(key))
		if readError == nil {
			dst, copyError := item.ValueCopy(data)
			if copyError == nil {
				data = dst
				return nil
			}
			return copyError
		}
		return readError
	})
	return data, err
}

func (l *LocalDB) ListKeys() []string {
	keys := make([]string, 0)
	_ = l.db.View(func(txn *badger.Txn) error {

		options := badger.DefaultIteratorOptions
		options.PrefetchValues = false
		iterator := txn.NewIterator(options)

		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			item := iterator.Item()
			keys = append(keys, string(item.Key()))
		}
		return nil
	})
	return keys
}

func (l *LocalDB) Remove(key string) error {
	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

func (l *LocalDB) NewSeq(memberName string) error {
	if _, ok := sequences[memberName]; !ok {
		if seq, err := l.db.GetSequence([]byte(memberName), 1); err == nil {
			sequences[memberName] = seq
			return nil
		} else {
			return err
		}
	}
	log.Printf("sequence key %s already exists", memberName)
	return nil
}

func (l *LocalDB) NextSeq(memberName string) (uint64, error) {
	if seq, ok := sequences[memberName]; ok {
		if next, err := seq.Next(); err != nil {
			return 0, fmt.Errorf("Unable to obtain the next sequence")
		} else {
			return next, nil
		}
	} else {
		if err := l.NewSeq(memberName); err == nil {
			if next, err := sequences[memberName].Next(); err == nil {
				return next, nil
			} else {
				return 0, fmt.Errorf("Unable to obtain the next sequence")
			}
		} else {
			return 0, fmt.Errorf("Unable to obtain the next sequence")
		}
	}
}

func (l *LocalDB) Close() {
	_ = l.db.Close()
}
