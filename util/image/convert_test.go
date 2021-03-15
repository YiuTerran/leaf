package image

import (
	"testing"
)

//把500k的图像压缩到50k
const src = "data:image/png;base64,iVBORw0KGgoAAAANSUhE"

func TestCompress(t *testing.T) {
	typ, data := SplitBase64HeaderData(src)
	t.Logf("%v, %v", typ, data)
}
