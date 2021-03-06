package setutil

import (
	"database/sql/driver"
	"encoding/json"
)

//如果需要通用泛型，可以使用github.com/deckarep/golang-set这个包
//非goroutine安全
type Int32Set struct {
	set map[int32]struct{}
}

//自定义序列化
func (d Int32Set) Value() (driver.Value, error) {
	return json.Marshal(d.ToArray())
}

func (d *Int32Set) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	var tmp []int32
	if err := json.Unmarshal(src.([]byte), &tmp); err != nil {
		return err
	}
	*d = *NewInt32Set(tmp...)
	return nil
}

func NewInt32Set(items ...int32) *Int32Set {
	d := &Int32Set{
		set: make(map[int32]struct{}, len(items)),
	}
	for _, item := range items {
		d.set[item] = struct{}{}
	}
	return d
}

func (d *Int32Set) Add(items ...int32) *Int32Set {
	for _, item := range items {
		d.set[item] = struct{}{}
	}
	return d
}

func (d *Int32Set) Remove(items ...int32) *Int32Set {
	for _, item := range items {
		delete(d.set, item)
	}
	return d
}

func (d *Int32Set) Contains(items ...int32) bool {
	var ok bool
	for _, item := range items {
		if _, ok = d.set[item]; !ok {
			return false
		}
	}
	return true
}

func (d *Int32Set) Size() int {
	return len(d.set)
}

//交集
func (d *Int32Set) Intersect(other *Int32Set) *Int32Set {
	result := NewInt32Set()
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
func (d *Int32Set) Union(other *Int32Set) *Int32Set {
	result := NewInt32Set()
	for k, v := range d.set {
		result.set[k] = v
	}
	for k, v := range other.set {
		result.set[k] = v
	}
	return result
}

//差集
func (d *Int32Set) Difference(other *Int32Set) *Int32Set {
	result := NewInt32Set()
	for k := range d.set {
		if !other.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (d *Int32Set) ToArray() []int32 {
	result := make([]int32, 0, d.Size())
	for k := range d.set {
		result = append(result, k)
	}
	return result
}
