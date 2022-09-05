package musicAPI

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var NeteaseServer string
var myHttpClient = &http.Client{Timeout: time.Second*10}

type Song struct {
	ID string			// ID
	Name string			// 歌名
	Singer string		// 歌手
	AlbumName string	// 专辑名
	AlbumPic string		// 专辑图片链接
	Alia string			// 主题曲、插曲
	Audio string		// 音频链接
	Time string			// 时长
	Size string			// 大小
	Flac string			// flac格式链接，仅咪咕音乐
	Lyric string		// 歌词
}

// GetLyricByID 获取单个歌词
func GetLyricByID(id string) string {
	uuu := fmt.Sprintf("http://%s/lyric?id=%s",NeteaseServer,id)
	rrr,err := http.Get(uuu)
	if err != nil {
		log.Println("歌词获取失败：", err)
		return ""
	}
	defer rrr.Body.Close()
	bbb,err := io.ReadAll(rrr.Body)
	if err != nil {
		log.Println("歌词获取失败：", err)
		return ""
	}
	var v LyricInfo
	err = json.Unmarshal(bbb, &v)
	if err != nil {
		log.Println("歌词获取失败：", err)
		return ""
	}
	return v.Lrc.Lyric
}

func NeteaseAPI(kw string) []Song {
	if kw == "" {
		log.Println("kw为空！")
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	limit := 30
	var R []Song
	var result = make(map[string]Song)
	// 搜索
	u := fmt.Sprintf("http://%s/cloudsearch?limit=%d&keywords=%s", NeteaseServer, limit, url.QueryEscape(kw))
	r,err := myHttpClient.Get(u)
	if err != nil {
		log.Println("myHttpClient.Get:", err)
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	if r.StatusCode != 200 {
		log.Println("r.StatusCode:", r.StatusCode)
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	defer r.Body.Close()
	b,err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("io.ReadAll:", err)
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	var searchDate NeteaseSearchInfo
	err = json.Unmarshal(b, &searchDate)			// 负数解析到int报错，不影响
	if err != nil {
		log.Println("json.Unmarshal:", err)
	}

	var IDS string
	for _,v := range searchDate.Result.Songs {
		id := fmt.Sprintf("%d", v.ID)
		name := v.Name
		var singer string
		if len(v.Ar) >=2 {
			singer = v.Ar[0].Name + "/" + v.Ar[1].Name
		}else if len(v.Ar) == 1 {
			singer = v.Ar[0].Name
		}
		albumName := v.Al.Name
		albumPic := v.Al.PicURL
		var alia string
		if len(v.Alia) > 0 {
			alia = v.Alia[0]
		}
		var audio,time,size,flac,lyric string
		IDS += id + ","
		result[id] = Song{id,name,singer,albumName,albumPic,alia, audio, time, size, flac, lyric}
	}

	// 获取音频信息
	if len(IDS) == 0 {
		log.Println("len(IDS)=0")
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	uu := fmt.Sprintf("http://%s/song/url?id=%s",NeteaseServer,IDS[:len(IDS)-1])
	rr,err := myHttpClient.Get(uu)
	if err != nil {
		log.Println("myHttpClient.Get:", err)
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	defer rr.Body.Close()
	bb,err := io.ReadAll(rr.Body)
	if err != nil {
		log.Println("io.ReadAll:", err)
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	var audioInfo NeteaseAudioInfo
	err = json.Unmarshal(bb, &audioInfo)
	if err != nil {
		log.Println(audioInfo)
		log.Println(string(bb))
		log.Println("json.Unmarshal:", err)
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	for _,vv := range audioInfo.Data {
		id := fmt.Sprintf("%d", vv.ID)
		t := result[id]
		t.Audio = vv.URL
		t.Time = fmt.Sprintf("%dm%ds", vv.Time/1000/60, vv.Time/1000%60)
		t.Size = fmt.Sprintf("%.1fM", float64(vv.Size)/1024/1024)
		result[id] = t

		// 过滤
		if t.Time == "0m30s" || t.Time == "0m0s" {
			delete(result, id)
		}
	}

	// 构造结果
	for _,v := range result {
		R = append(R, v)
	}
	return R
}

// DownloadMusic 下载歌曲
func DownloadMusic(url, path string) error  {
	r,err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http.Get:%s", err.Error())
	}
	if r.StatusCode != 200 {
		return fmt.Errorf("r.StatusCode:%d", r.StatusCode)
	}
	defer r.Body.Close()
	b,err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll:%s", err.Error())
	}
	err = os.WriteFile(path, b, 0755)
	if err != nil {
		return fmt.Errorf("os.WriteFile:%s", err.Error())
	}
	return nil
}

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

type NeteaseAudioInfo struct {
	Data []struct {
		ID                 int         `json:"id"`
		URL                string      `json:"url"`
		Br                 int         `json:"br"`
		Size               int         `json:"size"`
		Md5                string      `json:"md5"`
		Code               int         `json:"code"`
		Expi               int         `json:"expi"`
		Type               string      `json:"type"`
		Gain               float64     `json:"gain"`
		Fee                int         `json:"fee"`
		Uf                 interface{} `json:"uf"`
		Payed              int         `json:"payed"`
		Flag               int         `json:"flag"`
		CanExtend          bool        `json:"canExtend"`
		FreeTrialInfo      interface{} `json:"freeTrialInfo"`
		Level              string      `json:"level"`
		EncodeType         string      `json:"encodeType"`
		FreeTrialPrivilege struct {
			ResConsumable  bool        `json:"resConsumable"`
			UserConsumable bool        `json:"userConsumable"`
			ListenType     interface{} `json:"listenType"`
		} `json:"freeTrialPrivilege"`
		FreeTimeTrialPrivilege struct {
			ResConsumable  bool `json:"resConsumable"`
			UserConsumable bool `json:"userConsumable"`
			Type           int  `json:"type"`
			RemainTime     int  `json:"remainTime"`
		} `json:"freeTimeTrialPrivilege"`
		URLSource   int         `json:"urlSource"`
		RightSource int         `json:"rightSource"`
		PodcastCtrp interface{} `json:"podcastCtrp"`
		EffectTypes interface{} `json:"effectTypes"`
		Time        int         `json:"time"`
	} `json:"data"`
	Code int `json:"code"`
}

type NeteaseSearchInfo struct {
	Result struct {
		SearchQcReminder interface{} `json:"searchQcReminder"`
		Songs            []struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
			Pst  int    `json:"pst"`
			T    int    `json:"t"`
			Ar   []struct {
				ID    int           `json:"id"`
				Name  string        `json:"name"`
				Tns   []interface{} `json:"tns"`
				Alias []string      `json:"alias"`
				Alia  []string      `json:"alia"`
			} `json:"ar"`
			Alia []string    `json:"alia"`
			Pop  int         `json:"pop"`
			St   int         `json:"st"`
			Rt   interface{} `json:"rt"`
			Fee  int         `json:"fee"`
			V    int         `json:"v"`
			Crbt interface{} `json:"crbt"`
			Cf   string      `json:"cf"`
			Al   struct {
				ID     int           `json:"id"`
				Name   string        `json:"name"`
				PicURL string        `json:"picUrl"`
				Tns    []interface{} `json:"tns"`
				PicStr string        `json:"pic_str"`
				Pic    int64         `json:"pic"`
			} `json:"al"`
			Dt int `json:"dt"`
			H  struct {
				Br   int `json:"br"`
				Fid  int `json:"fid"`
				Size int `json:"size"`
				Vd   int `json:"vd"`
				Sr   int `json:"sr"`
			} `json:"h"`
			M struct {
				Br   int `json:"br"`
				Fid  int `json:"fid"`
				Size int `json:"size"`
				Vd   int `json:"vd"`
				Sr   int `json:"sr"`
			} `json:"m"`
			L struct {
				Br   int `json:"br"`
				Fid  int `json:"fid"`
				Size int `json:"size"`
				Vd   int `json:"vd"`
				Sr   int `json:"sr"`
			} `json:"l"`
			Sq struct {
				Br   int `json:"br"`
				Fid  int `json:"fid"`
				Size int `json:"size"`
				Vd   int `json:"vd"`
				Sr   int `json:"sr"`
			} `json:"sq"`
			Hr                   interface{}   `json:"hr"`
			A                    interface{}   `json:"a"`
			Cd                   string        `json:"cd"`
			No                   int           `json:"no"`
			RtURL                interface{}   `json:"rtUrl"`
			Ftype                int           `json:"ftype"`
			RtUrls               []interface{} `json:"rtUrls"`
			DjID                 int           `json:"djId"`
			Copyright            int           `json:"copyright"`
			SID                  int           `json:"s_id"`
			Mark                 int           `json:"mark"`
			OriginCoverType      int           `json:"originCoverType"`
			OriginSongSimpleData interface{}   `json:"originSongSimpleData"`
			TagPicList           interface{}   `json:"tagPicList"`
			ResourceState        bool          `json:"resourceState"`
			Version              int           `json:"version"`
			SongJumpInfo         interface{}   `json:"songJumpInfo"`
			EntertainmentTags    interface{}   `json:"entertainmentTags"`
			Single               int           `json:"single"`
			NoCopyrightRcmd      interface{}   `json:"noCopyrightRcmd"`
			Rtype                int           `json:"rtype"`
			Rurl                 interface{}   `json:"rurl"`
			Mst                  int           `json:"mst"`
			Cp                   int           `json:"cp"`
			Mv                   int           `json:"mv"`
			PublishTime          int64         `json:"publishTime"`
			Privilege            struct {
				ID                 int         `json:"id"`
				Fee                int         `json:"fee"`
				Payed              int         `json:"payed"`
				St                 int         `json:"st"`
				Pl                 int         `json:"pl"`
				Dl                 int         `json:"dl"`
				Sp                 int         `json:"sp"`
				Cp                 int         `json:"cp"`
				Subp               int         `json:"subp"`
				Cs                 bool        `json:"cs"`
				Maxbr              int         `json:"maxbr"`
				Fl                 int         `json:"fl"`
				Toast              bool        `json:"toast"`
				Flag               int         `json:"flag"`
				PreSell            bool        `json:"preSell"`
				PlayMaxbr          int         `json:"playMaxbr"`
				DownloadMaxbr      int         `json:"downloadMaxbr"`
				MaxBrLevel         string      `json:"maxBrLevel"`
				PlayMaxBrLevel     string      `json:"playMaxBrLevel"`
				DownloadMaxBrLevel string      `json:"downloadMaxBrLevel"`
				PlLevel            string      `json:"plLevel"`
				DlLevel            string      `json:"dlLevel"`
				FlLevel            string      `json:"flLevel"`
				Rscl               interface{} `json:"rscl"`
				FreeTrialPrivilege struct {
					ResConsumable  bool        `json:"resConsumable"`
					UserConsumable bool        `json:"userConsumable"`
					ListenType     interface{} `json:"listenType"`
				} `json:"freeTrialPrivilege"`
				ChargeInfoList []struct {
					Rate          int         `json:"rate"`
					ChargeURL     interface{} `json:"chargeUrl"`
					ChargeMessage interface{} `json:"chargeMessage"`
					ChargeType    int         `json:"chargeType"`
				} `json:"chargeInfoList"`
			} `json:"privilege"`
			Tns []string `json:"tns,omitempty"`
		} `json:"songs"`
		SongCount int `json:"songCount"`
	} `json:"result"`
	Code int `json:"code"`
}