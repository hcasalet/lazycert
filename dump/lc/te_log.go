package lc

import (
	"log"

	badger "github.com/dgraph-io/badger/v3"
)

type TELog struct {
	db *badger.DB
}

func NewTELog() *TELog {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}
	// Your code here…
	t := &TELog{}
	t.db = db

	return t
}

func (t *TELog) CertifyLog(index int32, certificate Certificate) {

}

func (t *TELog) UpdateLeader(termID int32, config LeaderConfig) {

}

func (t *TELog) Close() {
	t.db.Close()
}
