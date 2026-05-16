package middleware

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// tokenParamRe は クエリ文字列中の token= パラメータにマッチする
// 先頭または & の後に続く token= と、その値部分（次の & まで）を捕捉する
var tokenParamRe = regexp.MustCompile(`((?:^|&)token=)[^&]*`)

// maskTokenInPath はパスのクエリ文字列に含まれる token= の値を *** に置換する
func maskTokenInPath(path string) string {
	idx := strings.IndexByte(path, '?')
	if idx < 0 {
		return path
	}
	query := path[idx+1:]
	masked := tokenParamRe.ReplaceAllString(query, "${1}***")
	return path[:idx+1] + masked
}

// LoggerWithTokenMask は token クエリパラメータの値をマスクしてログ出力する
// gin.Logger() の代替として使用する
func LoggerWithTokenMask() gin.HandlerFunc {
	return loggerWithTokenMaskTo(gin.DefaultWriter)
}

func loggerWithTokenMaskTo(out io.Writer) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[GIN] %v | %3d | %13v | %15s | %-7s %s\n%s",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.StatusCode,
				param.Latency,
				param.ClientIP,
				param.Method,
				maskTokenInPath(param.Path),
				param.ErrorMessage,
			)
		},
		Output: out,
	})
}
