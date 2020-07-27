package setutil

import (
	"database/sql/driver"
	"encoding/json"
)

type StringSet struct {
	set map[string]struct{}
}

//自定义序列化
func (d StringSet) Value() (driver.Value, error) {
	return json.Marshal(d.ToArray())
}

func (d *StringSet) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	var tmp []string
	if err := json.Unmarshal(src.([]byte), &tmp); err != nil {
		return err
	}
	*d = *NewStringSet(tmp...)
	return nil
}

func NewStringSet(items ...string) *StringSet {
	d := &StringSet{
		set: make(map[string]struct{}, len(items)),
	}
	for _, item := range items {
		d.set[item] = struct{}{}
	}
	return d
}

func (d *StringSet) Add(items ...string) *StringSet {
	for _, item := range items {
		d.set[item] = struct{}{}
	}
	return d
}

func (d *StringSet) Remove(items ...string) *StringSet {
	for _, item := range items {
		delete(d.set, item)
	}
	return d
}

func (d *StringSet) Contains(items ...string) bool {
	var ok bool
	for _, item := range items {
		if _, ok = d.set[item]; !ok {
			return false
		}
	}
	return true
}

func (d *StringSet) Size() int {
	return len(d.set)
}

//交集
func (d *StringSet) Intersect(other *StringSet) *StringSet {
	result := NewStringSet()
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
func (d *StringSet) Union(other *StringSet) *StringSet {
	result := NewStringSet()
	for k, v := range d.set {
		result.set[k] = v
	}
	for k, v := range other.set {
		result.set[k] = v
	}
	return result
}

//差集
func (d *StringSet) Difference(other *StringSet) *StringSet {
	result := NewStringSet()
	for k := range d.set {
		if !other.Contains(k) {
			result.Add(k)
		}
	}
	return result
}

func (d *StringSet) ToArray() []string {
	result := make([]string, 0, d.Size())
	for k := range d.set {
		result = append(result, k)
	}
	return result
}
