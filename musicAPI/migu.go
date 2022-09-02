package musicAPI

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sync"
)

var miguServer = "39.101.203.25:3400"

func MiguAPI(kw string) []Song {
	var R []Song
	var result = make(map[string]Song)
	u := fmt.Sprintf("http://%s/search/?keyword=%s", miguServer, url.QueryEscape(kw))
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
		alia := ""
		audio := v.URL
		time := ""
		size := ""
		flac := ""
		result[id] = Song{id,name,singer,albumName,albumPic,alia, audio, time, size, flac}
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			uu := fmt.Sprintf("http://%s/song/?cid=%s", miguServer, id)
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
			t := result[data2.Data.Cid]
			t.Audio = data2.Data.Num320
			t.Time = fmt.Sprintf("%dm%ds", data2.Data.Duration/60,data2.Data.Duration%60)
			t.Flac = data2.Data.Flac
			result[data2.Data.Cid] = t
		}(id)
	}
	wg.Wait()
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


