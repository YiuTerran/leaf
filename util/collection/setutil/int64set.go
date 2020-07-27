package setutil

import (
	"database/sql/driver"
	"encoding/json"
)

type Int64Set struct {
	set map[int64]struct{}
}

//自定义序列化
func (d Int64Set) Value() (driver.Value, error) {
	return json.Marshal(d.ToArray())
}

func (d *Int64Set) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	var tmp []int64
	if err := json.Unmarshal(src.([]byte), &tmp); err != nil {
		return err
	}
	*d = *NewInt64Set(tmp...)
	return nil
}

func NewInt64Set(items ...int64) *Int64Set {
	d := &Int64Set{
		set: make(map[int64]struct{}, len(items)),
	}
	for _, item := range items {
		d.set[item] = struct{}{}
	}
	return d
}

func (d *Int64Set) Add(items ...int64) *Int64Set {
	for _, item := range items {
		d.set[item] = struct{}{}
	}
	return d
}

func (d *Int64Set) Remove(items ...int64) *Int64Set {
	for _, item := range items {
		delete(d.set, item)
	}
	return d
}

func (d *Int64Set) Contains(items ...int64) bool {
	var ok bool
	for _, item := range items {
		if _, ok = d.set[item]; !ok {
			return false
		}
	}
	return true
}

func (d *Int64Set) Size() int {
	return len(d.set)
}

//交集
func (d *Int64Set) Intersect(other *Int64Set) *Int64Set {
	result := NewInt64Set()
	//遍历较小的那个
	toRange, another := d.set, other
	if d.Size() > other.Size() {
		toRange, another = other.set, d
	}
	for k := range toRange {
		if another.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

//并集
func (d *Int64Set) Union(other *Int64Set) *Int64Set {
	result := NewInt64Set()
	for k, v := range d.set {
		result.set[k] = v
	}
	for k, v := range other.set {
		result.set[k] = v
	}
	return result
}

//差集
func (d *Int64Set) Difference(other *Int64Set) *Int64Set {
	result := NewInt64Set()
	for k := range d.set {
		if !other.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (d *Int64Set) ToArray() []int64 {
	result := make([]int64, 0, d.Size())
	for k := range d.set {
		result = append(result, k)
	}
	return result
}
