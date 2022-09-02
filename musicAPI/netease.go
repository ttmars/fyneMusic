package musicAPI

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// 本地虚拟机docker部署
//var neteaseServer = "192.168.66.102:3000"

// 通过Vercel部署，无需服务器！！！
// https://github.com/Binaryify/NeteaseCloudMusicApi
// https://vercel.com/ttmars/netease-cloud-music-api
//var neteaseServer = "netease-cloud-music-api-orcin-beta.vercel.app"		// 搜索结果比较少，有缺失。部署的分支有问题？还要开VPN访问？
var neteaseServer = "neteaseapi.youthsweet.com"

var myHttpClient = &http.Client{Timeout: time.Second*5}

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
}

func test() {
	kw := "林俊杰"
	r := NeteaseAPI(kw)
	fmt.Println(len(r))
	for _,v := range r {
		fmt.Println(v)
	}
}

func NeteaseAPI(kw string) []Song {
	var R = make([]Song, 0, 30)
	var r = make(map[string]Song)
	var n int
	var err error
	r,n,err = Netease(kw, 100)
	log.Println(n,err)
	if err != nil {
		log.Println(err)
		return []Song{{ID:"27731362", Name: "服务器错误!!!", Singer: "服务器错误!!!"}}
	}
	for _,v := range r {
		R = append(R, v)
	}
	if len(R) > 30 {
		R = R[:30]
	}
	return R
}

func Netease(kw string, limit int) (map[string]Song, int, error) {
	var result = make(map[string]Song)
	if kw == "" || limit < 1 {
		return result,len(result),errors.New("fuck")
	}
	// 搜索
	u := fmt.Sprintf("http://%s/cloudsearch?limit=%d&keywords=%s", neteaseServer, limit, url.QueryEscape(kw))
	r,err := myHttpClient.Get(u)
	if err != nil {
		return result,len(result),err
	}
	defer r.Body.Close()
	b,err := io.ReadAll(r.Body)
	if err != nil {
		return result,len(result),err
	}
	var searchDate NeteaseSearchInfo
	_ = json.Unmarshal(b, &searchDate)			// 负数解析到int报错，不影响
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
		var audio,time,size,flac string
		IDS += id + ","
		result[id] = Song{id,name,singer,albumName,albumPic,alia, audio, time, size, flac}
	}

	// 获取音频信息
	if len(IDS) == 0 {
		return result,len(result),errors.New("fuck")
	}
	uu := fmt.Sprintf("http://%s/song/url?id=%s",neteaseServer,IDS[:len(IDS)-1])
	rr,err := myHttpClient.Get(uu)
	if err != nil {
		return result,len(result),err
	}
	defer rr.Body.Close()
	bb,err := io.ReadAll(rr.Body)
	if err != nil {
		return result,len(result),err
	}
	var audioInfo NeteaseAudioInfo
	err = json.Unmarshal(bb, &audioInfo)
	if err != nil {
		return result,len(result),err
	}
	for _,vv := range audioInfo.Data {
		id := fmt.Sprintf("%d", vv.ID)
		t := result[id]
		t.Audio = vv.URL
		t.Time = fmt.Sprintf("%dm%ds", vv.Time/1000/60, vv.Time/1000%60)
		t.Size = fmt.Sprintf("%.1fM", float64(vv.Size)/1024/1024)
		result[id] = t
	}

	// 过滤试听歌曲
	for k,v := range result {
		if v.Time == "0m30s" || v.Time == "0m0s" {
			delete(result, k)
		}
	}

	return result,len(result),nil
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