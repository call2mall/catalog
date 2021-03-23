package wevikabot

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"time"
)

type WeViKaBot struct {
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (ua WeViKaBot) Header() (value string) {
	bytes := make([]byte, 10)
	for i := 0; i < 10; i++ {
		bytes[i] = byte(rand.Intn(99))
	}
	unique := fmt.Sprintf("%x", sha256.Sum256(bytes))[0:16]

	value = fmt.Sprintf("Mozilla/5.0 (compatible; WeViKaBot/1.0; +%s)", unique)

	return
}
