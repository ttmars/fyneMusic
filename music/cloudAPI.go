package music

import (
	"encoding/json"
	"io"
	"net/http"
)

func CloudAPI(kw string) []Song {
	url := "http://8.138.217.221:3800/cloud/api?keyword=" + kw
	resp, err := http.Get(url)
	if err != nil {
		return []Song{{ID: "27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return []Song{{ID: "27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	var v []Song
	err = json.Unmarshal(b, &v)
	if err != nil {
		return []Song{{ID: "27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	return v
}
