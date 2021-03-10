package jsonutil

import (
	"database/sql/driver"
	"time"

	"github.com/YiuTerran/leaf/util/tz"
	"github.com/araddon/dateparse"
)

type Time struct {
	time.Time
}

func (t *Time) string() string {
	return t.Format(tz.FullFormat)
}

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	realT, err := dateparse.ParseLocal(string(data))
	if err != nil {
		return err
	}
	(*t).Time = realT
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(t.Format(tz.FullFormat)), nil
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
