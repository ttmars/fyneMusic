package musicAPI

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sync"
)

var MiguServer string

func MiguAPI(kw string) []Song {
	if kw == "" {
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	var R []Song
	var result = make(map[string]Song)
	u := fmt.Sprintf("http://%s/search/?keyword=%s", MiguServer, url.QueryEscape(kw))
	r,err := myHttpClient.Get(u)
	if err != nil {
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	defer r.Body.Close()
	b,err := io.ReadAll(r.Body)
	if err != nil {
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	var data1 MiguSearchInfo
	err = json.Unmarshal(b, &data1)
	if err != nil {
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	for _,v := range data1.Data.List {
		id := v.Cid
		name := v.Name
		var singer string
		if len(v.Artists) == 1 {
			singer = v.Artists[0].Name
		}else if len(v.Artists) == 2 {
			singer = v.Artists[0].Name + v.Artists[1].Name
		}
		albumName := v.Album.Name
		albumPic := v.Album.PicURL
		audio := v.URL
		var alia, time, size, flac, lyric string
		result[id] = Song{id,name,singer,albumName,albumPic,alia, audio, time, size, flac, lyric}
	}

	// 音频链接
	var r2 [][]string
	var wg sync.WaitGroup
	for k,_ := range result{
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			uu := fmt.Sprintf("http://%s/song/?cid=%s", MiguServer, id)
			rr,err := myHttpClient.Get(uu)
			if err != nil {
				return
			}
			defer rr.Body.Close()
			bb,err := io.ReadAll(rr.Body)
			if err != nil {
				return
			}
			var data2 MiguAudioInfo
			err = json.Unmarshal(bb, &data2)
			if err != nil {
				return
			}
			r2 = append(r2, []string{id, data2.Data.Num320, fmt.Sprintf("%dm%ds", data2.Data.Duration/60,data2.Data.Duration%60), data2.Data.Flac, data2.Data.Lyric})
		}(k)
	}
	wg.Wait()

	// 合并结果
	for i:=0;i<len(r2);i++{
		t := result[r2[i][0]]
		t.Audio = r2[i][1]
		t.Time = r2[i][2]
		t.Flac = r2[i][3]
		t.Lyric = r2[i][4]
		result[r2[i][0]] = t
	}

	// 返回结果
	for _,v := range result {
		R = append(R, v)
	}
	return R
}


type MiguSearchInfo struct {
	Result int `json:"result"`
	Data   struct {
		List []struct {
			Name  string `json:"name"`
			ID    string `json:"id"`
			Cid   string `json:"cid"`
			MvID  string `json:"mvId"`
			URL   string `json:"url"`
			Album struct {
				PicURL string `json:"picUrl"`
				Name   string `json:"name"`
				ID     string `json:"id"`
			} `json:"album"`
			Artists []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"list"`
		Total int `json:"total"`
	} `json:"data"`
}

type MiguAudioInfo struct {
	Result int `json:"result"`
	Data   struct {
		Num128  string `json:"128"`
		Num320  string `json:"320"`
		ID      string `json:"id"`
		Cid     string `json:"cid"`
		Name    string `json:"name"`
		Artists []struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			NameSpelling string `json:"nameSpelling"`
		} `json:"artists"`
		Album struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"album"`
		Duration  int    `json:"duration"`
		MvID      string `json:"mvId"`
		MvCid     string `json:"mvCid"`
		PicURL    string `json:"picUrl"`
		BigPicURL string `json:"bigPicUrl"`
		Flac      string `json:"flac"`
		Lyric     string `json:"lyric"`
	} `json:"data"`
}


