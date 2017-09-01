package main

import (
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"bytes"
	"encoding/json"
	"strings"
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

type YearStats struct {
	CupSize int `json:"cupSize"`
	NumOfTime int `json:"numOfTime"`
	NumOfPeople int `json:"numOfPeople"`
	NumOfNew int `json:"numOfNew"`
}

func yearFilter(ctx *gin.Context) {
	num := ctx.Param("num")
	year, err := strconv.ParseInt(num, 10, 64)
	if err != nil || year < 2007 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, Reply{Msg: "年份日期错误"})
		return
	}

	ctx.Set("year", year)
	ctx.Next()
}

func getDailyLog(value []byte) *Body {
	body := &Body{
		Names:   make([]string, 0),
		Tags:    make([]string, 0),
		Comment: "",
		CupSize: 0,
	}
	if err := json.Unmarshal(value, &body); err != nil {
		return nil
	}

	return body
}

func updateYearStats(stats *YearStats, body *Body) {
	stats.CupSize += body.CupSize
	stats.NumOfTime += len(body.Names)
	for _, tags := range body.Tags {
		if strings.Contains(tags, "新人") {
			stats.NumOfNew += 1
		}
	}
}

func (env *Env) getYear(ctx *gin.Context) {
	year := ctx.GetInt64("year")
	loc := ctx.GetString("loc")
	yearStats := YearStats{0, 0, 0, 0}
	nameMap := map[string]bool{}
	err := env.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte(loc)).Cursor()

		min := []byte(strconv.FormatInt(year, 10) + "-01-01")
		max := []byte(strconv.FormatInt(year+1, 10) + "-01-01")

		for k, v := cursor.Seek(min); k != nil && bytes.Compare(k, max) < 0; k, v = cursor.Next() {
			dailyLog := getDailyLog(v)
			if dailyLog == nil {
				continue
			}
			updateYearStats(&yearStats, dailyLog)
			for _, name := range dailyLog.Names {
				nameMap[name] = true
			}
		}

		return nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, Reply{Msg: "读取年份错误"})
		return
	}

	yearStats.NumOfPeople = len(nameMap)
	ctx.JSON(http.StatusOK, yearStats)
}
