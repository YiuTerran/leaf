package jsonutil

import (
	"encoding/json"
	"testing"
)

var (
	str = []byte(`{"fruit": [{"name": "apple", "color": 1},{"name":"banana", "color":2}], "owner": null, "ref": 999}`)
)

func TestLoadsKey(t *testing.T) {
	var obj JsonObject
	if json.Unmarshal(str, &obj) != nil {
		t.Error("fail to load")
		return
	}
	fs, err := obj.GetJsonArray("fruit")
	if err != nil {
		t.Error(err)
	}
	if len(fs) != 2 {
		t.Errorf("fruit length error!, %s\n", fs)
	}
	ff, err := fs.ToObjectArray()
	if err != nil {
		t.Errorf("fail to convert to obj array:%s\n", err)
	}
	for _, f := range ff {
		t.Log(f)
	}
}
