# 电子签到

> go get github.com/steambap/signin

[![Build Status](https://travis-ci.org/steambap/signin.svg)](https://travis-ci.org/steambap/signin)

## API

### get /log 获取日志

参数：

- date 时间
- loc 心栈编号

示例：/log?date=2017-09-06&loc=11

### post /log 保存日志

参数：

- date 时间
- loc 心栈编号

示例：/log?date=2017-09-06&loc=11

### get /loc/:loc/year/:num 获取年数据

参数：

- loc 心栈编号
- num 年份

示例： /loc/11/year/2017

### get /loc/:loc 获取所有日期

参数：

- loc 心栈编号

示例： /loc/11

### get /loc/:loc/week/:day 获取周数据

参数：

- loc 心栈编号
- day 日期

示例： /loc/11/week/2017-09-25

# License

所有代码的所有权和最终解释权属于北京龙泉寺
