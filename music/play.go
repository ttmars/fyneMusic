package music

import (
	"fmt"
	"fyne.io/fyne/v2/theme"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/patrickmn/go-cache"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var MyPlayer = NewPlayer()

type Player struct {
	PlayMode int             // 播放模式，1:单曲循环	2:随机播放
	Speed float64            // 播放的倍数(0.5~2)
	LabelLimitLength int     // 歌曲列表宽度限制(与窗口大小关联)
	SearchCache *cache.Cache // 搜索缓存，防止重复请求
	MusicChan chan Song      // 存放歌曲信息，播放器监听，有数据则播放
	DoneChan chan bool       // 播放器将播放完毕后的信号传入此通道

	KeyWord         string            // 当前搜索关键字
	SearchAPI       string            // 当前搜索API
	CurrentSong     Song              // 当前播放歌曲信息
	CurrentLyricMap map[string]string // 当前歌曲的歌词信息，动态解析，动态更新
	PlayList        []Song            // 当前播放列表信息
	DownloadPath    string            // 下载路径
	MiguServer      string            // 咪咕服务器
	NeteaseServer   string            // 网易云服务器
}

func NewPlayer() *Player {
	curDir,_ := os.Getwd()
	downloadPath := curDir + "\\download"
	return &Player{
		PlayMode: 2,
		Speed: 1,
		LabelLimitLength: 15,
		SearchCache: cache.New(20*time.Minute, 5*time.Minute),
		MusicChan: make(chan Song, 1),
		DoneChan: make(chan bool, 1),
		DownloadPath: downloadPath,
		MiguServer: "39.101.203.25:3400",
		NeteaseServer: "43.138.26.158:3000",
	}
}

// RandomPlay 随机播放线程
func (p *Player)RandomPlay()  {
	for {
		select {
		case <-p.DoneChan:
			rand.Seed(time.Now().Unix())
			randomNum := rand.Intn(len(p.PlayList))
			p.MusicChan <- p.PlayList[randomNum]
		}
	}
}

// InitPlayList 异步初始化线程
func (p *Player)InitPlayList()  {
	p.PlayList = NeteaseAPI("抖音")
	p.KeyWord = "抖音"
	p.SearchAPI = "网易云"
	ml.Refresh()
	p.DoneChan <- true		// 启动后随机播放
}

// PlayMusic 异步播放器线程
func (p *Player)PlayMusic()  {
	for {
		select{
		case song := <-p.MusicChan:
			speaker.Clear()
			r,err := http.Get(song.Audio)
			if err != nil || r.StatusCode != 200 {
				//dialog.ShowInformation("播放失败", "版权限制\n"+p.CurrentSong.Name+"_"+p.CurrentSong.Singer+".mp3", W)
				log.Println("自动刷新数据!", p.SearchAPI, p.KeyWord)
				go searchFunc(p.SearchAPI, p.KeyWord)
				break
			}
			defer r.Body.Close()
			// 写入文件
			b,err := io.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
				break
			}
			f,err := os.Create(os.TempDir() + "\\tmp.mp3")
			if err != nil {
				log.Println(err)
				break
			}
			f.Write(b)
			f.Seek(0,0)
			streamer, musicFormat, err = mp3.Decode(f)		// 原始音频流
			if err != nil {
				log.Println(err)
				break
			}
			defer streamer.Close()

			musicStreamer = beep.ResampleRatio(4, 1, streamer)		// 控制播放速度
			musicStreamer.SetRatio(p.Speed)
			ctrl = &beep.Ctrl{Streamer: musicStreamer, Paused: false}			// 暂  停/播  放

			_ = speaker.Init(musicFormat.SampleRate, musicFormat.SampleRate.N(time.Second/10))
			speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
				if p.PlayMode == 1 {
					p.MusicChan <- song			// 单曲循环
				}else{
					p.DoneChan <- true			// 随机播放
				}
			})))
			// 更新状态
			p.CurrentSong = song
			mp.PlayerLabel.SetText(song.Name + "_" + song.Singer + ".mp3")
			mp.PlayerLabel.Refresh()
			mp.PlayerPause.SetIcon(theme.MediaPauseIcon())
			mp.PlayerPause.SetText("暂  停")
			mp.PlayerLyric.SetText("")
			mp.PlayerLyric.Refresh()
			p.CurrentLyricMap = make(map[string]string)
			if song.Lyric != "" {
				p.CurrentLyricMap = p.parseLyric(song.Lyric)
			}else {
				go func() {
					start := time.Now()
					p.CurrentLyricMap = p.parseLyric(GetLyricByID(song.ID))
					log.Println("单条歌词请求耗时：", time.Since(start))
				}()
			}
		}
	}
}


// UpdateProgressLabel 异步更新线程
func (p *Player)UpdateProgressLabel()  {
	for {
		select {
		case <-time.After(time.Millisecond * 500):
			if streamer == nil {
				break
			}
			// 歌词刷新
			cur := musicFormat.SampleRate.D(streamer.Position()).Round(time.Second)
			keyTime := fmt.Sprintf("%02d:%02d", int(cur.Seconds())/60, int(cur.Seconds())%60)
			if lyric,ok := p.CurrentLyricMap[keyTime];ok{
				mp.PlayerLyric.SetText(lyric)
			}
			mp.PlayerLyric.Refresh()

			// 进度条刷新
			total := fmt.Sprintf("%02d:%02d", p.CurrentSong.Time/60, p.CurrentSong.Time%60)
			mp.PlayerEndTime.SetText(total)
			mp.PlayerEndTime.Refresh()
			mp.PlayerStartTime.SetText(keyTime)
			mp.PlayerStartTime.Refresh()
			mp.PlayerProgress.Max = float64(streamer.Len())
			mp.PlayerProgress.SetValue(float64(streamer.Position()))
			mp.PlayerProgress.Refresh()
		}
	}
}

// 解析歌词
func (p *Player)parseLyric(s string) map[string]string {
	m := make(map[string]string)
	t := strings.Split(s, "\n")
	for _,v := range t {
		if len(v) > 10{
			v1 := v[1:6]
			v2 := v[10:]
			if v2[0] == ']' {
				v2 = v2[1:]
			}
			m[v1] = v2
		}
	}
	if p.CurrentSong.AlbumName != "" {
		m["00:05"] = p.CurrentSong.AlbumName
	}else if p.CurrentSong.Singer != ""{
		m["00:05"] = p.CurrentSong.Singer
	}else {
		m["00:05"] = p.CurrentSong.Name
	}
	return m
}