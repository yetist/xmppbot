package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/UssieApp/graphics-go/graphics"
	"image"
	_ "image/gif"  //必须import，否则会出现：unknown format，其余类似
	_ "image/jpeg" //必须import，否则会出现：unknown format，其余类似
	"image/png"
	"strings"
)

func GetBase64Image(body []byte, width, height int) string {
	src, _, err := image.Decode(strings.NewReader(string(body))) //解码图片
	if err != nil {
		return ""
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	err = graphics.Scale(dst, src) //缩小图片
	if err != nil {
		return ""
	}
	buf := bytes.NewBuffer([]byte{})
	err = png.Encode(buf, dst) //编码图片
	if err != nil {
		return ""
	}
	e64 := base64.StdEncoding
	maxEncLen := e64.EncodedLen(buf.Len())
	encBuf := make([]byte, maxEncLen)
	e64.Encode(encBuf, buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", string(encBuf))
}
