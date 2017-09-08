package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/steambap/signin/middleware"
	"log"
	"net/http"
	"strconv"
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

func fmtBytes(size int) string {
	if size > (1 << 10) {
		return strconv.FormatFloat(
			float64(size)/float64(1<<10), 'f', 3, 64) + "KB"
	}

	return strconv.FormatInt(int64(size), 10) + "B"
}

func xzLogger(ctx *gin.Context) {
	ctx.Next()
	loc := ctx.GetString("loc")
	if loc == "" {
		return
	}
	log.Printf("-->%6s|%6s", loc, fmtBytes(ctx.Writer.Size()))
}

func locationFilter(ctx *gin.Context) {
	// 11 默认天津南开心栈
	locKey := ctx.DefaultQuery("loc", "11")
	if locVal, ok := bucketMap[locKey]; ok {
		ctx.Set("loc", locVal)
	} else {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "心栈位置错误"})
		return
	}
	// Test Bucket has only one date
	if locKey == "0" {
		ctx.Set("dateKey", "2017-03-01")
	}

	ctx.Next()
}

func (env *Env) getDaily(ctx *gin.Context) {
	var body = Body{
		Names:   make([]string, 0),
		Tags:    make([]string, 0),
		Comment: "",
		CupSize: -1,
	}
	dateKey := ctx.GetString("dateKey")
	loc := ctx.GetString("loc")
	err := env.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(loc))

		if err := json.Unmarshal(bucket.Get([]byte(dateKey)), &body); err != nil {
			return nil
		} else {
			return nil
		}
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
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "数据解析错误"})
		return
	}

	ctx.Set("body", val)
	ctx.Next()
}

func (env *Env) putDaily(ctx *gin.Context) {
	body := ctx.Keys["body"].([]byte)
	dateKey := ctx.GetString("dateKey")
	loc := ctx.GetString("loc")
	err := env.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(loc))

		err := bucket.Put([]byte(dateKey), body)

		return err
	})
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, Reply{Msg: "写入数据库错误"})
		return
	}
	ctx.JSON(http.StatusOK, Reply{Msg: "OK"})
}

func main() {
	db := startBoltDb(DB_NAME)
	defer db.Close()

	env := &Env{db: db}

	router := gin.Default()
	router.Use(xzLogger)
	router.Use(middleware.Compress())
	router.Use(static.Serve("/", static.LocalFile("./public", true)))

	router.GET("/log", dateFilter, locationFilter, env.getDaily)
	router.POST("/log", dateFilter, locationFilter, marshalBody, env.putDaily)

	router.GET("/loc/:loc/year/:num", yearFilter, locationParamFilter, env.getYear)
	router.GET("/loc/:loc", locationParamFilter, env.scanBucket)
	router.GET("/loc/:loc/week/:day", dayParamFilter, locationParamFilter, env.handleWeek)

	router.Run(":8900")
}

func assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
