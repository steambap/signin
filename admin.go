package main

import (
	"bytes"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (env *Env) scanBucket(ctx *gin.Context) {
	loc := ctx.GetString("loc")
	prefix := ctx.DefaultQuery("prefix", "2017")
	result := map[string]string{}

	err := env.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte(loc)).Cursor()

		bPrefix := []byte(prefix)
		for k, v := cursor.Seek(bPrefix); k != nil && bytes.HasPrefix(k, bPrefix); k, v = cursor.Next() {
			result[string(k)] = string(v)
		}
		return nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Reply{Msg: "扫描数据库出错"})
	} else {
		ctx.JSON(http.StatusOK, result)
	}
}
