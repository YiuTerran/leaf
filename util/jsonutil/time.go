package jsonutil

import (
	"database/sql/driver"
	"time"

	"github.com/YiuTerran/leaf/log"
	"github.com/YiuTerran/leaf/util/tz"
	"github.com/araddon/dateparse"
)

type Time struct {
	time.Time
}

func (t *Time) String() string {
	return t.Format(tz.FullFormat)
}

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	if len(data) <= 2 {
		return nil
	}
	data = data[1 : len(data)-1] //去除双引号
	realT, err := dateparse.ParseLocal(string(data))
	if err != nil {
		log.Warn("fail to parse %s to time", string(data))
		return err
	}
	(*t).Time = realT
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

//自定义序列化
func (t *Time) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return t.Time, nil
}

func (t *Time) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	switch src.(type) {
	case time.Time:
		(*t).Time = src.(time.Time)
	case int64: //假设都用毫秒
		(*t).Time = time.Unix(src.(int64)/1000, 0)
	case string:
		(*t).Time, _ = dateparse.ParseLocal(src.(string))
	}
	return nil
}
