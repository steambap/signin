package main

import (
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func locationParamFilter(ctx *gin.Context) {
	// 11 默认天津南开心栈
	locKey := ctx.Param("loc")
	if locVal, ok := bucketMap[locKey]; ok {
		ctx.Set("loc", locVal)
	} else {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "心栈位置错误"})
		return
	}

	ctx.Next()
}

func (env *Env) scanBucket(ctx *gin.Context) {
	loc := ctx.GetString("loc")
	result := []string{}

	err := env.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(loc))

		b.ForEach(func(k, _ []byte) error {
			result = append(result, string(k))
			return nil
		})
		return nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Reply{Msg: "扫描数据库出错"})
	} else {
		ctx.JSON(http.StatusOK, result)
	}
}
