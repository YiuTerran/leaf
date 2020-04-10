package image

import (
	"bufio"
	"bytes"
	"encoding/base64"

	"github.com/disintegration/imaging"
)

//压缩图片到指定格式
//保持宽高比
func Compress(data []byte, width int, format imaging.Format) ([]byte, error) {
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	resized := imaging.Resize(img, width, 0, imaging.Lanczos)
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	if err = imaging.Encode(writer, resized, format); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ConvertToBase64(img []byte) string {
	return base64.StdEncoding.EncodeToString(img)
}

func ConvertToImage(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
