package main

import (
	"github.com/boltdb/bolt"
)

const DB_NAME = "xz.db"

func startBoltDb() *bolt.DB {
	db, err := bolt.Open(DB_NAME, 0600, nil)
	assert(err)

	for _, bucketName := range bucketMap {
		err = db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(bucketName))

			return err
		})
		assert(err)
	}

	return db
}
