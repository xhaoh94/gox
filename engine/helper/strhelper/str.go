package strhelper

import (
	"encoding/json"
	"hash/crc32"
	"strconv"
)

// ValToString 获取变量的字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func ValToString(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		if b, err := json.Marshal(value); err == nil {
			key = string(b)
		} else {
			key = ""
		}
	}
	return key
}

// StringToInt 字符串转int
func StringToInt(str string) int {
	if num, err := strconv.Atoi(str); err != nil {
		return 0
	} else {
		return num
	}
}

// StringToHash 字符串转为32位整形哈希
func StringToHash(s string) (hash uint32) {

	hash = crc32.ChecksumIEEE([]byte(s))
	if hash >= 0 {
		return hash
	}
	if -hash >= 0 {
		return -hash
	}

	for _, c := range s {
		ch := uint32(c)
		hash = hash + ((hash) << 5) + ch + (ch << 7)
	}
	return
}
