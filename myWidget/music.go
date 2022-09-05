package myWidget

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyneMusic/musicAPI"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/patrickmn/go-cache"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// 计算当前播放进度
var streamer beep.StreamSeekCloser
var musicFormat beep.Format

// 缓存
var musicCache = cache.New(20*time.Minute, 5*time.Minute)

// 按钮
var searchSubmit *widget.Button			// 搜索按钮，防止重复点击
var playButton *widget.Button			// 播放按钮
var flacDownloadButton *widget.Button	// 下载按钮
var downloadButton *widget.Button		// 下载按钮
var pauseButton *widget.Button			// 暂停按钮
var playNext *widget.Button				// 下一首

// 标签
var lyricLabel *widget.Label			// 歌词栏
var lyricMap map[string]string
var playerLabel *widget.Hyperlink
var speedLabel *widget.Label			// 倍速标签，控制样式
var progressLabel *widget.Label
var speedSlider *widget.Slider			// 倍速滑动条，控制样式
var labelLength = 15					// 歌曲名称长度
var MusicData []musicAPI.Song			// 歌曲信息
var CurrentMusic musicAPI.Song			// 当前播放歌曲
var MusicDataContainer []fyne.CanvasObject				// 装载歌曲信息的容器
var BasePath,_ = filepath.Abs(".") // 下载路径
var SavePath string

var musicStreamer *beep.Resampler		// 控制播放速度
var ctrl *beep.Ctrl						// 控制暂停
var musicCh = make(chan string, 1)		// 控制播放
var doneCh = make(chan bool, 1)			// 控制随机播放

var W fyne.Window
var A fyne.App
// RandomPlay 随机播放线程
func RandomPlay()  {
	doneCh <- true		// 启动后，随机播放歌曲
	for {
		select {
		case <-doneCh:
			rand.Seed(time.Now().Unix())
			randomNum := rand.Intn(len(MusicData))
			musicCh <- MusicData[randomNum].Audio
			CurrentMusic = MusicData[randomNum]
			playerLabel.SetText(MusicData[randomNum].Name + "_" + MusicData[randomNum].Singer + ".mp3")
			playerLabel.Refresh()
			if pauseButton != nil {
				pauseButton.SetIcon(theme.MediaPauseIcon())
				pauseButton.SetText("暂停")
			}
			if speedLabel != nil {
				speedLabel.SetText("1.0倍速")
				speedSlider.SetValue(1)
			}
		}
	}
}

// PlayMusic 播放器线程
func PlayMusic()  {
	var link string
	for {
		select{
		case link = <-musicCh:
			r,err := http.Get(link)
			if err != nil || r.StatusCode != 200 {
				dialog.ShowInformation("播放失败", "链接失效或版权限制\n"+CurrentMusic.Name+"_"+CurrentMusic.Singer+".mp3", W)
				break
			}
			defer r.Body.Close()

			streamer, musicFormat, err = mp3.Decode(r.Body)		// 原始音频流
			if err != nil {
				dialog.ShowInformation("播放失败", "链接失效或版权限制\n"+CurrentMusic.Name+"_"+CurrentMusic.Singer+".mp3", W)
				break
			}
			defer streamer.Close()

			musicStreamer = beep.ResampleRatio(4, 1, streamer)		// 控制播放速度
			ctrl = &beep.Ctrl{Streamer: musicStreamer, Paused: false}			// 暂停/播放

			_ = speaker.Init(musicFormat.SampleRate, musicFormat.SampleRate.N(time.Second/10))
			speaker.Clear()
			speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
				doneCh <- true
			})))
			lyricLabel.SetText("")
			lyricLabel.Refresh()
			if CurrentMusic.Lyric == "" {
				go func() {
					start := time.Now()
					lyricMap = parseLyric(musicAPI.GetLyricByID(CurrentMusic.ID))
					log.Println("单条歌词请求耗时：", time.Since(start))
				}()
			}else {
				lyricMap = parseLyric(CurrentMusic.Lyric)
			}
		}
	}
}

// UpdateProgressLabel 动态刷新歌词、播放时间
func UpdateProgressLabel()  {
	for {
		select {
		case <-time.After(time.Second):
			cur := musicFormat.SampleRate.D(streamer.Position()).Round(time.Second)

			keyTime := fmt.Sprintf("%02d:%02d", int(cur.Seconds())/60, int(cur.Seconds())%60)
			if lyric,ok := lyricMap[keyTime];ok{
				lyricLabel.SetText(lyric)
			}
			lyricLabel.Refresh()

			total := CurrentMusic.Time
			v := fmt.Sprintf("%s/%s",cur,total)
			progressLabel.SetText(v)
			progressLabel.Refresh()
		}
	}
}

// 解析歌词
func parseLyric(s string) map[string]string {
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
	if CurrentMusic.AlbumName != "" {
		m["00:01"] = CurrentMusic.AlbumName
	}else if CurrentMusic.Singer != ""{
		m["00:01"] = CurrentMusic.Singer
	}else {
		m["00:01"] = CurrentMusic.Name
	}
	return m
}

// 播放器
func createPlayer() fyne.CanvasObject {
	playNext = widget.NewButtonWithIcon("下一首", theme.MediaSkipNextIcon(), func() {
		playNext.Disable()
		defer playNext.Enable()
		pauseButton.SetIcon(theme.MediaPauseIcon())
		pauseButton.SetText("暂停")
		if speedLabel != nil {
			speedLabel.SetText("1.0倍速")
			speedSlider.SetValue(1)
		}
		doneCh <- true
	})
	playerLabel = widget.NewHyperlink("hello", nil)
	playerLabel.OnTapped = func() {
		if lyricLabel.Visible() {
			lyricLabel.Hide()
		}else{
			lyricLabel.Show()
			lyricLabel.Refresh()
		}
	}

	speedSlider = widget.NewSlider(0.5,1.5)
	speedSlider.SetValue(1)
	speedSlider.Step = 0.1
	speedSlider.OnChanged = func(f float64) {
		musicStreamer.SetRatio(f)
		speedLabel.SetText(fmt.Sprintf("%.1f倍速", f))
	}
	speedLabel = widget.NewLabel("1.0倍速")
	progressLabel = widget.NewLabel("0/0")
	pauseButton = widget.NewButtonWithIcon("暂停", theme.MediaPauseIcon(), func() {
		if ctrl != nil {
			ctrl.Paused = !ctrl.Paused
			if ctrl.Paused {
				pauseButton.SetIcon(theme.MediaPlayIcon())
				pauseButton.SetText("播放")
			}else{
				pauseButton.SetIcon(theme.MediaPauseIcon())
				pauseButton.SetText("暂停")
			}
		}
	})
	lyricLabel = widget.NewLabel("")
	lyricContainer := container.NewCenter(lyricLabel)
	sp := container.NewBorder(nil,nil,speedLabel,progressLabel,speedSlider)
	player := container.NewHBox(pauseButton, playNext)
	player = container.NewBorder(nil, sp, playerLabel, player,lyricContainer)
	return player
}

// MusicMerge 整合部件
func MusicMerge(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	return container.NewBorder(searchWidget(myApp, parent), createPlayer(), nil, nil, createMusicList(myApp, parent))
}

// 搜索组件
func searchWidget(myApp fyne.App, parent fyne.Window)fyne.CanvasObject  {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("遇见萤火")

	searchEngine := widget.NewSelectEntry([]string{"咪咕", "网易云"})
	searchEngine.SetText("咪咕")

	searchSubmit = widget.NewButtonWithIcon("搜索",theme.SearchIcon(), func() {
		if searchEntry.Text == "" {
			return
		}
		searchSubmit.Disable()
		defer searchSubmit.Enable()

		// 重新请求数据、创建组件并刷新
		if searchEngine.Text == "网易云" {
			if x, found := musicCache.Get("网易云"+searchEntry.Text); found {
				MusicData = x.([]musicAPI.Song)
			}else{
				cur := time.Now()
				t := musicAPI.NeteaseAPI(searchEntry.Text)
				log.Println("网易云耗时：", time.Since(cur))
				if len(t) == 1 {
					dialog.ShowInformation("搜索失败", "网易云API服务器出错.", W)
					return
				}else{
					MusicData = t
					musicCache.SetDefault("网易云"+searchEntry.Text, MusicData)
				}
			}
		}else{
			if x, found := musicCache.Get("咪咕"+searchEntry.Text); found {
				MusicData = x.([]musicAPI.Song)
			}else{
				cur := time.Now()
				t := musicAPI.MiguAPI(searchEntry.Text)
				log.Println("咪咕耗时：", time.Since(cur))
				if len(t) == 1 {
					dialog.ShowInformation("搜索失败", "咪咕API服务器出错.", W)
					return
				}else{
					MusicData = t
					musicCache.SetDefault("咪咕"+searchEntry.Text, MusicData)
				}
			}
		}

		for i,song := range MusicData {
			if i < len(MusicDataContainer) {
				MusicDataContainer[i] = createOneMusic(song, myApp, parent)
				MusicDataContainer[i].Refresh()
			}
		}

		doneCh <- true		// 搜索后自动随机播放
	})
	search := container.NewBorder(nil,nil,nil,searchSubmit, searchEngine)
	search = container.NewGridWithColumns(5, searchEntry, search)

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
	titleLabel := widget.NewLabel(cutString(song.Name, labelLength))
	singerLabel := widget.NewLabel(cutString(song.Singer, labelLength))
	u,_ := url.Parse(song.AlbumPic)
	albumLabel := widget.NewHyperlink(cutString(song.AlbumName, labelLength), u)
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
		CurrentMusic = song
		pauseButton.SetIcon(theme.MediaPauseIcon())
		pauseButton.SetText("暂停")
		playerLabel.SetText(song.Name + "_" + song.Singer + ".mp3")
		playerLabel.Refresh()
		speedLabel.SetText("1.0倍速")
		speedSlider.SetValue(1)
		musicCh <- song.Audio
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
func createMusicList(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	MusicDataContainer = make([]fyne.CanvasObject,0,30)		// 固定30
	MusicData = musicAPI.MiguAPI("周杰伦")
	//MusicData = musicAPI.NeteaseAPI("王力宏")
	length := len(MusicData)
	if length == 0 {
		for i:=0;i<30;i++{
			MusicDataContainer = append(MusicDataContainer, widget.NewSeparator())
		}
	} else if length < 30 {
		for _,song := range MusicData {
			MusicDataContainer = append(MusicDataContainer, createOneMusic(song, myApp, parent))
		}
		for i:=0;i<(30-length);i++{
			MusicDataContainer = append(MusicDataContainer, widget.NewSeparator())
		}
	} else {
		for k,song := range MusicData {
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