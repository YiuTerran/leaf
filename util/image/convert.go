package image

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"strings"

	"github.com/disintegration/imaging"
)

const base64Header = ";base64,"

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

func RemoveImgBase64Header(img string) string {
	if idx := strings.Index(img, base64Header); idx >= 0 {
		img = strings.TrimPrefix(img, img[:idx+8]) //8是header的长度
	}
	return img
}

func AddImgBase64Header(img string) string {
	if strings.HasPrefix(img, "data:image") {
		return img
	}
	return "data:image/jpeg;base64," + img
}

func ConvertToBase64(img []byte) string {
	return AddImgBase64Header(base64.StdEncoding.EncodeToString(img))
}

func ConvertToImage(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(RemoveImgBase64Header(data))
}
