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
	var wg sync.WaitGroup
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
		var v0,v1,v2,v3,v4 []string
		result[id] = Song{id,name,singer,albumName,albumPic,alia, audio, time, size, flac, lyric}
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
			v0 = append(v0, id)
			v1 = append(v1, data2.Data.Num320)
			v2 = append(v2, fmt.Sprintf("%dm%ds", data2.Data.Duration/60,data2.Data.Duration%60))
			v3 = append(v3, data2.Data.Flac)
			v4 = append(v4, data2.Data.Lyric)
		}(id)
		wg.Wait()

		for i:=0;i<len(v0)-1;i++{
			t := result[v0[i]]
			t.Audio = v1[i]
			t.Time = v2[i]
			t.Flac = v3[i]
			t.Lyric = v4[i]
			result[v0[i]] = t
		}
	}

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


