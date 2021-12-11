package helper

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func FillStruct_Interface(v interface{}, result interface{}) {
	data := v.(map[string]interface{})
	FillStruct(data, result)
}

func FillStruct(data map[string]interface{}, result interface{}) {
	if err := mapstructure.Decode(data, &result); err != nil {
		fmt.Println(err)
	}
}

func MD5Encode(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}

func MD5Check(input string, encrypted string) bool {
	return strings.EqualFold(MD5Encode(input), encrypted)
}

func GetUniqueRandom(m map[int]bool, max int) int {
	for {
		i, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
		if err != nil {
			fmt.Println("Pickup Random number Fail")
		}

		num := int(i.Int64())
		if !m[num] {
			m[num] = true
			return num
		}
	}
}
