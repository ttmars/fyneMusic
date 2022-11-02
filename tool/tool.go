package tool

import (
	"math/rand"
	"os"
	"time"
)

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// GetRandomNM 从[N,M]中获取随机数
func GetRandomNM(N,M int) int  {
	if M<N {
		return 0
	}
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(M-N+1)
	return n+N
}
