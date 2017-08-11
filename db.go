package main

import (
	"github.com/boltdb/bolt"
)

const DB_NAME = "xz.db"

var BUCKET_NAME = []byte("天津")

func startBoltDb() *bolt.DB {
	db, err := bolt.Open(DB_NAME, 0600, nil)
	assert(err)

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(BUCKET_NAME)

		return err
	})
	assert(err)

	return db
}
