package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type Env struct {
	db *bolt.DB
}

type Body struct {
	Names   []string `json:"names" binding:"required"`
	Tags    []string `json:"tags" binding:"required"`
	Comment string   `json:"comment"`
	CupSize int      `json:"cup_size"`
}

type Reply struct {
	Msg string `json:"msg"`
}

func tryFixDate(dateArr []string) []string {
	if len(dateArr[1]) == 1 {
		dateArr[1] = "0" + dateArr[1]
	}
	if len(dateArr[2]) == 1 {
		dateArr[2] = "0" + dateArr[2]
	}

	return dateArr
}

func dateFilter(ctx *gin.Context) {
	dateStr := ctx.Query("date")
	dateArr := strings.Split(dateStr, "-")
	if len(dateArr) != 3 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "日期格式错误"})
		return
	}
	dateArr = tryFixDate(dateArr)
	date, err := time.Parse("2006-01-02", strings.Join(dateArr, "-"))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "日期格式错误"})
		return
	}

	ctx.Set("dateKey", date.Format("2006-01-02"))
	ctx.Next()
}

func (env *Env) getDaily(ctx *gin.Context) {
	var body Body
	dateKey := ctx.GetString("dateKey")
	err := env.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BUCKET_NAME)

		err := json.Unmarshal(bucket.Get([]byte(dateKey)), &body)
		return err
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, Reply{Msg: "读取数据库错误"})
		return
	}
	ctx.JSON(http.StatusOK, body)
}

func marshalBody(ctx *gin.Context) {
	var body Body
	if err := ctx.Bind(&body); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "数据格式错误"})
		return
	}

	val, err := json.Marshal(body)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "数据格式错误"})
		return
	}

	ctx.Set("body", string(val))
	ctx.Next()
}

func (env *Env) putDaily(ctx *gin.Context) {
	body := []byte(ctx.GetString("body"))
	dateKey := ctx.GetString("dateKey")
	err := env.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(BUCKET_NAME)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(dateKey), body)

		return err
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, Reply{Msg: "写入数据库错误"})
		return
	}
	ctx.JSON(http.StatusOK, Reply{Msg: "OK"})
}

func main() {
	db := startBoltDb()
	defer db.Close()

	env := &Env{db: db}

	router := gin.Default()

	router.GET("/log", dateFilter, env.getDaily)
	router.POST("/log", dateFilter, marshalBody, env.putDaily)

	router.Run(":8900")
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
