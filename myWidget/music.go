package myWidget

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyneMusic/musicAPI"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/patrickmn/go-cache"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Player struct {
	PlayMode int 						// 播放模式，1:单曲循环	2:随机播放
	Speed float64						// 播放的倍数(0.5~2)
	LabelLimitLength int				// 歌曲列表宽度限制(与窗口大小关联)
	SearchCache *cache.Cache			// 搜索缓存，防止重复请求
	MusicChan chan musicAPI.Song		// 存放歌曲信息，播放器监听，有数据则播放
	DoneChan chan bool					// 播放器将播放完毕后的信号传入此通道

	KeyWord string						// 当前搜索关键字
	SearchAPI string					// 当前搜索API
	CurrentSong musicAPI.Song			// 当前播放歌曲信息
	CurrentLyricMap map[string]string	// 当前歌曲的歌词信息，动态解析，动态更新
	PlayList []musicAPI.Song			// 当前播放列表信息
}

func NewPlayer() *Player {
	return &Player{
		PlayMode: 2,
		Speed: 1,
		LabelLimitLength: 15,
		SearchCache: cache.New(20*time.Minute, 5*time.Minute),
		MusicChan: make(chan musicAPI.Song, 1),
		DoneChan: make(chan bool, 1),
	}
}
var MyPlayer = NewPlayer()

var W fyne.Window
var A fyne.App
var BasePath,_ = filepath.Abs(".")
var SavePath string

var streamer beep.StreamSeekCloser
var musicFormat beep.Format
var musicStreamer *beep.Resampler		// 控制播放速度
var ctrl *beep.Ctrl						// 控制暂停

var searchSubmit *widget.Button			// 搜索按钮，防止重复点击
var playButton *widget.Button			// 播放按钮
var flacDownloadButton *widget.Button	// 下载按钮
var downloadButton *widget.Button		// 下载按钮
var pauseButton *widget.Button			// 暂停按钮
var playNext *widget.Button				// 下一曲
var progressLeft *widget.Label
var progressRight *widget.Label
var line3 *fyne.Container
var lyricLabel *widget.Label			// 歌词栏
var playerLabel *widget.Hyperlink
var speedLabel *widget.Hyperlink		// 倍速标签，控制样式
var progressSlider *widget.Slider
var speedSlider *widget.Slider							// 倍速滑动条，控制样式
var MusicDataContainer []fyne.CanvasObject				// 装载歌曲信息的容器

// InitPlayList 异步初始化
func (p *Player)InitPlayList()  {
	p.PlayList = musicAPI.MiguAPI("周杰伦")
	p.KeyWord = "周杰伦"
	p.SearchAPI = "咪咕"
	for i,song := range p.PlayList {
		if i < len(MusicDataContainer) {
			MusicDataContainer[i] = createOneMusic(song, A, W)
			MusicDataContainer[i].Refresh()
		}
	}
	p.DoneChan <- true		// 启动后随机播放
}

// RandomPlay 异步随机播放线程
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
			playerLabel.SetText(song.Name + "_" + song.Singer + ".mp3")
			playerLabel.Refresh()
			pauseButton.SetIcon(theme.MediaPauseIcon())
			pauseButton.SetText("暂  停")
			lyricLabel.SetText("")
			lyricLabel.Refresh()
			p.CurrentLyricMap = make(map[string]string)
			if song.Lyric != "" {
				p.CurrentLyricMap = p.parseLyric(song.Lyric)
			}else {
				go func() {
					start := time.Now()
					p.CurrentLyricMap = p.parseLyric(musicAPI.GetLyricByID(song.ID))
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
				lyricLabel.SetText(lyric)
			}
			lyricLabel.Refresh()

			// 进度条刷新
			total := fmt.Sprintf("%02d:%02d", p.CurrentSong.Time/60, p.CurrentSong.Time%60)
			progressRight.SetText(total)
			progressRight.Refresh()
			progressLeft.SetText(keyTime)
			progressLeft.Refresh()
			progressSlider.Max = float64(streamer.Len())
			progressSlider.SetValue(float64(streamer.Position()))
			progressSlider.Refresh()
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

// MusicMerge 整合部件
func MusicMerge(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	return container.NewBorder(searchWidget(), createPlayer(), nil, nil, createMusicList(myApp, parent))
}

// 播放器
func createPlayer() fyne.CanvasObject {
	// 第一行
	playerLabel = widget.NewHyperlink("", nil)
	playerLabel.OnTapped = func() {
		if lyricLabel.Visible() {
			lyricLabel.Hide()
		}else{
			lyricLabel.Show()
			lyricLabel.Refresh()
		}
	}
	lyricLabel = widget.NewLabel("")
	lyricContainer := container.NewCenter(lyricLabel)
	pauseButton = widget.NewButtonWithIcon("暂  停", theme.MediaPauseIcon(), func() {
		if ctrl != nil {
			ctrl.Paused = !ctrl.Paused
			if ctrl.Paused {
				pauseButton.SetIcon(theme.MediaPlayIcon())
				pauseButton.SetText("播  放")
			}else{
				pauseButton.SetIcon(theme.MediaPauseIcon())
				pauseButton.SetText("暂  停")
			}
		}
	})
	line1 := container.NewBorder(nil,nil,playerLabel,pauseButton, lyricContainer)

	// 第二行
	speedLabel = widget.NewHyperlink("1.0倍速", nil)
	speedLabel.OnTapped = func() {
		if line3.Visible() {
			line3.Hide()
		}else{
			line3.Show()
			line3.Refresh()
		}
	}
	progressLeft = widget.NewLabel("00:00")
	progressSlider = widget.NewSlider(0,10)
	progressSlider.SetValue(0)
	progressSlider.Step = 1
	progressSlider.OnChanged = func(f float64) {
		if streamer != nil {
			t := math.Abs(float64(streamer.Position())-f)
			if  t > 200000 {
				ctrl.Paused = true
				streamer.Seek(int(f))
				ctrl.Paused = false
			}
		}
	}
	progressRight = widget.NewLabel("00:00")
	playNext = widget.NewButtonWithIcon("下一曲", theme.MediaSkipNextIcon(), func() {
		playNext.Disable()
		defer playNext.Enable()
		MyPlayer.DoneChan <- true
	})
	line2 := container.NewBorder(nil,nil,progressLeft,progressRight,progressSlider)
	line2 = container.NewBorder(nil,nil,speedLabel,playNext, line2)

	// 第三行
	speedSliderLeftLabel := widget.NewLabel("0.5")
	speedSliderRightLabel := widget.NewLabel("2.0")
	speedSlider = widget.NewSlider(0.5,2.0)
	speedSlider.SetValue(1)
	speedSlider.Step = 0.1
	speedSlider.OnChanged = func(f float64) {
		MyPlayer.Speed = f
		musicStreamer.SetRatio(f)
		speedLabel.SetText(fmt.Sprintf("%.1f倍速", f))
		speedSliderLeftLabel.SetText(fmt.Sprintf("%.1f", f))
	}
	line3 = container.NewBorder(nil,nil,speedSliderLeftLabel,speedSliderRightLabel,speedSlider)
	line3.Hide()

	player := container.NewVBox(line1, line2)
	player = container.NewBorder(nil,line3,nil,nil,player)
	return player
}

// 搜索组件
func searchWidget()fyne.CanvasObject  {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("遇见萤火")

	searchEngine := widget.NewSelectEntry([]string{"网易云", "咪咕"})
	searchEngine.SetText("网易云")

	searchSubmit = widget.NewButtonWithIcon("搜索",theme.SearchIcon(), func() {
		searchFunc(searchEngine.Text, searchEntry.Text)
	})
	search := container.NewBorder(nil,nil,nil,searchSubmit, searchEngine)
	search = container.NewGridWithColumns(5, searchEntry, search)
	playModeCheck := widget.NewCheck("单曲循环", func(b bool) {
		if b {
			MyPlayer.PlayMode = 1
		}else {
			MyPlayer.PlayMode = 2
		}
	})
	search = container.NewBorder(nil,nil,nil, playModeCheck, search)

	// 音乐标签
	titleLabel := widget.NewLabelWithStyle("音乐标题", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	singerLabel := widget.NewLabel("歌手")
	linkLabel := widget.NewLabel("点播")
	albumLabel := widget.NewLabel("专辑")
	downloadLabel := widget.NewLabelWithStyle("标准", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	flacDownloadLabel := widget.NewLabelWithStyle("无损", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	down := container.NewGridWithColumns(2, downloadLabel, flacDownloadLabel)
	musicLabel := container.NewGridWithColumns(2, linkLabel, down)
	musicLabel = container.NewGridWithColumns(4, titleLabel, singerLabel,albumLabel,musicLabel)

	return container.NewVBox(search, musicLabel)
}

func searchFunc(eg, kw string)  {
	log.Println("搜索：", eg, kw)
	if kw == "" {
		return
	}
	MyPlayer.KeyWord = kw
	MyPlayer.SearchAPI = eg
	searchSubmit.Disable()
	defer searchSubmit.Enable()

	// 重新请求数据、创建组件并刷新
	if eg == "网易云" {
		if x, found := MyPlayer.SearchCache.Get("网易云"+kw); found {
			MyPlayer.PlayList = x.([]musicAPI.Song)
		}else{
			cur := time.Now()
			t := musicAPI.NeteaseAPI(kw)
			log.Println("网易云请求耗时：", time.Since(cur))
			if len(t) == 1 {
				dialog.ShowInformation("搜索失败", "网易云API服务器出错.", W)
				return
			}else{
				MyPlayer.PlayList = t
				MyPlayer.SearchCache.SetDefault("网易云"+kw, MyPlayer.PlayList)
			}
		}
	}else{
		if x, found := MyPlayer.SearchCache.Get("咪咕"+kw); found {
			MyPlayer.PlayList = x.([]musicAPI.Song)
		}else{
			cur := time.Now()
			t := musicAPI.MiguAPI(kw)
			log.Println("咪咕耗时：", time.Since(cur))
			if len(t) == 1 {
				dialog.ShowInformation("搜索失败", "咪咕API服务器出错.", W)
				return
			}else{
				MyPlayer.PlayList = t
				MyPlayer.SearchCache.SetDefault("咪咕"+kw, MyPlayer.PlayList)
			}
		}
	}

	for i,song := range MyPlayer.PlayList {
		if i < len(MusicDataContainer) {
			MusicDataContainer[i] = createOneMusic(song, A, W)
			MusicDataContainer[i].Refresh()
		}
	}

	MyPlayer.DoneChan <- true		// 搜索后自动随机播放
}

// 字符串长度裁剪
func cutString(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}else{
		return string(r[:n])
	}
}

func createOneMusic(song musicAPI.Song, myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	titleLabel := widget.NewLabel(cutString(song.Name, MyPlayer.LabelLimitLength))
	singerLabel := widget.NewLabel(cutString(song.Singer, MyPlayer.LabelLimitLength))
	u,_ := url.Parse(song.AlbumPic)
	albumLabel := widget.NewHyperlink(cutString(song.AlbumName, MyPlayer.LabelLimitLength), u)
	albumLabel.OnTapped = func() {
		w := myApp.NewWindow("image")
		w.CenterOnScreen()
		w.SetContent(CreateImage(song.AlbumPic))
		w.Resize(fyne.NewSize(400,400))
		w.Show()
	}
	playButton = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		playButton.Disable()
		defer playButton.Enable()
		MyPlayer.MusicChan <- song
	})
	downloadButton = widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {
		downloadButton.Disable()
		defer downloadButton.Enable()
		path := SavePath + "\\" + song.Name + "_" + song.Singer + ".mp3"
		err := musicAPI.DownloadMusic(song.Audio, path)
		if err != nil {
			dialog.ShowInformation("下载失败!", err.Error(), parent)
		}else{
			dialog.ShowInformation("下载成功!", path, parent)
		}
	})

	flacDownloadButton = widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {
		flacDownloadButton.Disable()
		defer flacDownloadButton.Enable()
		path := SavePath + "\\" + song.Name + "_" + song.Singer + ".flac"
		err := musicAPI.DownloadMusic(song.Flac, path)
		if err != nil {
			dialog.ShowInformation("下载失败!", err.Error(), parent)
		}else{
			dialog.ShowInformation("下载成功!", path, parent)
		}
	})
	if song.Flac == "" {
		flacDownloadButton.Disable()
	}

	down := container.NewGridWithColumns(2, downloadButton, flacDownloadButton)
	onemusic := container.NewGridWithColumns(2, playButton, down)
	onemusic = container.NewGridWithColumns(4, titleLabel, singerLabel,albumLabel,onemusic)
	return onemusic
}

// 歌曲列表组件
func createMusicList(myApp fyne.App, parent fyne.Window) *container.Scroll {
	MusicDataContainer = make([]fyne.CanvasObject,0,30)		// 固定30
	length := len(MyPlayer.PlayList)
	if length == 0 {
		for i:=0;i<30;i++{
			MusicDataContainer = append(MusicDataContainer, layout.NewSpacer())
		}
	} else if length < 30 {
		for _,song := range MyPlayer.PlayList {
			MusicDataContainer = append(MusicDataContainer, createOneMusic(song, myApp, parent))
		}
		for i:=0;i<(30-length);i++{
			MusicDataContainer = append(MusicDataContainer, widget.NewSeparator())
		}
	} else {
		for k,song := range MyPlayer.PlayList {
			if k < 30 {
				MusicDataContainer = append(MusicDataContainer, createOneMusic(song, myApp, parent))
			}
		}
	}
	return container.NewVScroll(container.NewVBox(MusicDataContainer...))
}

// CreateImage 创建一个图片
func CreateImage(pic string) *canvas.Image {
	r,err := http.Get(pic)
	if err != nil {
		return canvas.NewImageFromResource(theme.FyneLogo())
	}
	defer r.Body.Close()
	image := canvas.NewImageFromReader(r.Body, "jpg")
	//image.FillMode = canvas.ImageFillOriginal
	return image
}
