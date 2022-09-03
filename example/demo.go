package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type LyricInfo struct {
	Sgc       bool `json:"sgc"`
	Sfy       bool `json:"sfy"`
	Qfy       bool `json:"qfy"`
	LyricUser struct {
		ID       int    `json:"id"`
		Status   int    `json:"status"`
		Demand   int    `json:"demand"`
		Userid   int    `json:"userid"`
		Nickname string `json:"nickname"`
		Uptime   int64  `json:"uptime"`
	} `json:"lyricUser"`
	Lrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"lrc"`
	Klyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"klyric"`
	Tlyric struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"tlyric"`
	Romalrc struct {
		Version int    `json:"version"`
		Lyric   string `json:"lyric"`
	} `json:"romalrc"`
	Code int `json:"code"`
}

func main() {
	u := "http://192.168.66.102:3000/lyric?id=27731362"
	r,_ := http.Get(u)
	defer r.Body.Close()
	b,_ := io.ReadAll(r.Body)
	var v LyricInfo
	_ = json.Unmarshal(b, &v)
	fmt.Println(v.Lrc.Lyric)
}
