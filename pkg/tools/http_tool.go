package tools

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type RespDataType struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data"`
}

func RespData(c *gin.Context, code int, msg interface{}, data ...interface{}) {
	if len(data) == 0 {
		c.JSON(200,
			RespDataType{
				Code: code,
				Msg:  msg,
			})
	} else {
		c.JSON(200,
			RespDataType{
				Code: code,
				Msg:  msg,
				Data: data[0],
			})
	}

}

func Resp400(c *gin.Context, msg string, data ...interface{}) {
	c.JSON(http.StatusOK,
		RespDataType{
			Code: http.StatusBadRequest,
			Msg:  msg,
			Data: data,
		})
}

func Resp400BadRequest(c *gin.Context) {
	c.JSON(http.StatusOK,
		RespDataType{
			Code: http.StatusBadRequest,
			Msg:  "参数错误",
		})
}
func Resp500(c *gin.Context, msg string, data ...interface{}) {
	c.JSON(http.StatusOK,
		RespDataType{
			Code: http.StatusInternalServerError,
			Msg:  msg,
			Data: data})
}

func Resp401(ctx *gin.Context) {
	ctx.JSON(http.StatusUnauthorized,
		RespDataType{
			Code: http.StatusUnauthorized,
			Msg:  "not device ",
		})
}
