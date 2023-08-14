package tools

import (
	"go.uber.org/zap"
	"math/rand"
	"time"
)

// OperationIDGenerator 用时间戳生成一个ID
func OperationIDGenerator() int64 {
	return time.Now().UnixNano() + int64(rand.Uint32())
}

func HandlePanic(logger *zap.Logger, tag string) {
	// 未知原因的错误处理
	// 捕捉错误
	errRecover := recover()
	if errRecover == nil {
		// 没有错误
		return
	}

	//err = errRecover
	// 有错误 直接打印错误
	logger.Error(tag, zap.Any("err", errRecover))
}

func GetCrc8(buf []byte) (crc byte) {
	for i := 0; i < len(buf); i++ {
		crc ^= buf[i]
		for j := 8; j > 0; j-- {
			if crc&0x01 > 0 {
				crc = (crc >> 1) ^ 0x8c
			} else {
				crc = crc >> 1
			}
		}
	}

	return
}

func JenkinsHash(data []byte) int64 {
	var hash uint64
	//var i uint64

	for i := 0; i < len(data); i++ {
		hash = uint64(data[i])

		hash += hash << 10
		hash ^= hash >> 6

	}

	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15

	return int64(hash)

}
