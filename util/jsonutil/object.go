package jsonutil

import (
	"errors"

	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
)

//一个动态的json对象
//注意：json在Unmarshal到interface{}时，会把JsonNumber转成float64，除非使用UseNumber
//因此这里仅提供float64接口，其他数据类型外部转换
//如果json的是{type: 1, data: {}}这种格式，需要通过type解析具体的data，则推荐使用json.RawMessage来解析

type JsonObject map[string]interface{}

var (
	TypeError  = errors.New("type convert error")
	KeyError   = errors.New("key not exist")
	IndexError = errors.New("index not exist")
)

//自定义序列化
func (j JsonObject) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JsonObject) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	d := make(map[string]interface{})
	if err := json.Unmarshal(src.([]byte), &d); err != nil {
		return err
	}
	*j = d
	return nil
}

func (j JsonObject) GetInt(key string) (int, error) {
	if v, ok := j[key]; !ok {
		return 0, errors.New("key not exist")
	} else {
		switch v.(type) {
		case float64:
			return int(v.(float64)), nil
		case string:
			return strconv.Atoi(v.(string))
		default:
			return 0, errors.New("type error")
		}
	}
}

func (j JsonObject) GetInt64(key string) (int64, error) {
	if v, ok := j[key]; !ok {
		return 0, errors.New("key not exist")
	} else {
		switch v.(type) {
		case float64:
			return int64(v.(float64)), nil
		case string:
			return strconv.ParseInt(v.(string), 0, 64)
		default:
			return 0, errors.New("type error")
		}
	}
}

func (j JsonObject) GetString(key string) (string, error) {
	if v, ok := j[key]; !ok {
		return "", errors.New("key not exist")
	} else {
		switch v.(type) {
		case string:
			return v.(string), nil
		case float64:
			return fmt.Sprint(v), nil
		default:
			return "", errors.New("type error")
		}
	}
}

func (j JsonObject) GetStringDefault(key string, def string) string {
	if v, err := j.GetString(key); err != nil {
		return def
	} else {
		return v
	}
}

func (j JsonObject) GetIntDefault(key string, def int) int {
	v, err := j.GetInt(key)
	if err == nil {
		return v

	}
	return def
}

func (j JsonObject) GetInt64Default(key string, def int64) int64 {
	v, err := j.GetInt64(key)
	if err == nil {
		return v
	}
	return def
}

func (j JsonObject) HasKey(key string) bool {
	if _, ok := j[key]; ok {
		return true
	}
	return false
}

func (j JsonObject) HasNotNilKey(key string) bool {
	if tmp, ok := j[key]; ok {
		if tmp != nil {
			return true
		}
	}
	return false
}

func (j JsonObject) GetFloat64(key string) (float64, error) {
	var (
		tmp  interface{}
		resp float64
		ok   bool
	)
	if tmp, ok = j[key]; ok {
		if resp, ok = tmp.(float64); ok {
			return resp, nil
		}
		return 0, TypeError
	}
	return 0, KeyError
}

func (j JsonObject) GetFloat64Default(key string, defaultValue float64) float64 {
	var (
		tmp  interface{}
		resp float64
		ok   bool
	)
	if tmp, ok = j[key]; ok {
		if resp, ok = tmp.(float64); ok {
			return resp
		}
	}
	return defaultValue
}

func (j JsonObject) GetBool(key string) (bool, error) {
	var (
		tmp  interface{}
		resp bool
		ok   bool
	)
	if tmp, ok = j[key]; ok {
		if resp, ok = tmp.(bool); ok {
			return resp, nil
		}
		return false, TypeError
	}
	return false, KeyError
}

func (j JsonObject) GetBoolDefault(key string, defaultValue bool) bool {
	var (
		tmp  interface{}
		resp bool
		ok   bool
	)
	if tmp, ok = j[key]; ok {
		if resp, ok = tmp.(bool); ok {
			return resp
		}
	}
	return defaultValue
}

func (j JsonObject) GetJsonArray(key string) (JsonArray, error) {
	var (
		tmp  interface{}
		resp []interface{}
		ok   bool
	)
	if tmp, ok = j[key]; ok {
		if resp, ok = tmp.([]interface{}); ok {
			return resp, nil
		}
		return nil, TypeError
	}
	return nil, KeyError
}

func (j JsonObject) GetJsonObject(key string) (JsonObject, error) {
	var (
		tmp  interface{}
		resp map[string]interface{}
		ok   bool
	)
	if tmp, ok = j[key]; ok {
		if resp, ok = tmp.(map[string]interface{}); ok {
			return resp, nil
		}
		return nil, TypeError
	}
	return nil, KeyError
}
