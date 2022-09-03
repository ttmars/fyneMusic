package myWidget

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyneMusic/musicAPI"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"math/rand"
	"net/http"
	"path/filepath"
	"time"
)

var MusicData []musicAPI.Song							// 歌曲信息
var MusicDataContainer []fyne.CanvasObject				// 装载歌曲信息的容器
var BasePath,_ = filepath.Abs(".") // 下载路径
var musicLength = 30					// 歌曲名称长度
var pauseButton *widget.Button			// 暂停按钮，控制样式
var searchSubmit *widget.Button			// 搜索按钮，防止重复点击
var musicName = binding.NewString()		// 当前播放歌曲名称

var speedLabel *widget.Label			// 倍速标签，控制样式
var speedSlider *widget.Slider			// 倍速滑动条，控制样式
var musicStreamer *beep.Resampler		// 控制播放速度
var ctrl *beep.Ctrl						// 控制暂停

var musicCh = make(chan string, 1)		// 控制播放
var doneCh = make(chan bool, 1)			// 控制随机播放

var W fyne.Window
// RandomPlay 随机播放线程
func RandomPlay()  {
	doneCh <- true		// 启动后，随机播放歌曲
	for {
		select {
		case <-doneCh:
			rand.Seed(time.Now().Unix())
			if len(MusicData) == 0 {
				dialog.ShowInformation("搜索失败!", "请更换关键词！", W)
				break
			}
			randomNum := rand.Intn(len(MusicData))
			musicCh <- MusicData[randomNum].Audio
			musicName.Set("随机播放：" + MusicData[randomNum].Name + "_" + MusicData[randomNum].Singer + ".mp3")
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
				//log.Fatal(err)
				dialog.ShowInformation("链接失效!", "请重新搜索刷新数据！", W)
				break
			}
			defer r.Body.Close()

			streamer, musicFormat, err := mp3.Decode(r.Body)		// 原始音频流
			if err != nil {
				//log.Fatal(err)
				dialog.ShowInformation("链接失效!", "请重新搜索刷新数据！", W)
				break
			}
			defer streamer.Close()

			musicStreamer = beep.ResampleRatio(4, 1, streamer)		// 控制播放速度
			ctrl = &beep.Ctrl{Streamer: musicStreamer, Paused: false}			// 暂停/播放

			speaker.Init(musicFormat.SampleRate, musicFormat.SampleRate.N(time.Second/10))
			speaker.Clear()
			speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
				doneCh <- true
			})))
		}
	}
}

// 播放器
func createPlayer(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	playNext := widget.NewButtonWithIcon("下一首", theme.MediaSkipNextIcon(), func() {
		pauseButton.SetIcon(theme.MediaPauseIcon())
		pauseButton.SetText("暂停")
		if speedLabel != nil {
			speedLabel.SetText("1.0倍速")
			speedSlider.SetValue(1)
		}
		doneCh <- true
	})
	playerLabel := widget.NewLabelWithData(musicName)
	speedSlider = widget.NewSlider(0.5,1.5)
	speedSlider.SetValue(1)
	speedSlider.Step = 0.1
	speedSlider.OnChanged = func(f float64) {
		musicStreamer.SetRatio(f)
		speedLabel.SetText(fmt.Sprintf("%.1f倍速", f))
	}
	speedLabel = widget.NewLabel("1.0倍速")
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
	sp := container.NewBorder(nil, nil, speedLabel, nil, speedSlider)
	player := container.NewHBox(pauseButton, playNext)
	player = container.NewBorder(nil, sp, playerLabel, player)
	return player
}

// MusicMerge 整合部件
func MusicMerge(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	W = parent
	return container.NewBorder(searchWidget(myApp, parent), createPlayer(myApp, parent), nil, nil, createMusicList(myApp, parent))
}

// 搜索组件
func searchWidget(myApp fyne.App, parent fyne.Window)fyne.CanvasObject  {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("遇见萤火")

	searchEngine := widget.NewSelectEntry([]string{"网易云", "咪咕音乐"})
	searchEngine.SetText("网易云")

	searchSubmit = widget.NewButtonWithIcon("搜索",theme.SearchIcon(), func() {
		if searchEntry.Text == "" {
			return
		}
		searchSubmit.Disable()

		// 重新请求数据、创建组件并刷新
		if searchEngine.Text == "网易云" {
			MusicData = musicAPI.NeteaseAPI(searchEntry.Text)
		}else{
			MusicData = musicAPI.MiguAPI(searchEntry.Text)
		}

		for i,song := range MusicData {
			if i < len(MusicDataContainer) {
				MusicDataContainer[i] = createOneMusic(song, myApp, parent)
				MusicDataContainer[i].Refresh()
			}
		}

		doneCh <- true		// 搜索后自动随机播放
		searchSubmit.Enable()
	})
	search := container.NewBorder(nil,nil,nil,searchSubmit, searchEngine)
	search = container.NewGridWithColumns(5, searchEntry, search)

	// 音乐标签
	titleLabel := widget.NewLabelWithStyle("音乐标题", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	singerLabel := widget.NewLabelWithStyle("歌手", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	linkLabel := widget.NewLabelWithStyle("点播", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	downloadLabel := widget.NewLabelWithStyle("下载", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	musicLabel := container.NewGridWithColumns(2, linkLabel, downloadLabel)
	musicLabel = container.NewGridWithColumns(3, titleLabel, singerLabel, musicLabel)

	return container.NewVBox(search, musicLabel)
}

func createOneMusic(song musicAPI.Song, myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	r := []rune(song.Name)
	if len(r) > musicLength {
		r = r[:musicLength]
		song.Name = string(r)
	}
	titleLable := widget.NewLabel(song.Name)
	singerLable := widget.NewLabel(song.Singer)
	playButton := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		pauseButton.SetIcon(theme.MediaPauseIcon())
		pauseButton.SetText("暂停")
		musicName.Set("当前点播：" + song.Name + "_" + song.Singer + ".mp3")
		speedLabel.SetText("1.0倍速")
		speedSlider.SetValue(1)
		musicCh <- song.Audio
	})
	downloadButton := widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {
		path := myApp.Preferences().StringWithFallback("SongSavePath", BasePath)
		path = path + "\\" + song.Name + "_" + song.Singer + ".mp3"
		err := musicAPI.DownloadMusic(song.Audio, path)
		if err != nil {
			dialog.ShowInformation("下载失败!", err.Error(), parent)
		}else{
			dialog.ShowInformation("下载成功!", path, parent)
		}
	})

	onemusic := container.NewGridWithColumns(2, playButton, downloadButton)
	onemusic = container.NewGridWithColumns(3, titleLable, singerLable,onemusic)
	return onemusic
}

// 歌曲列表组件
func createMusicList(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	MusicDataContainer = make([]fyne.CanvasObject,0,30)		// 固定30
	MusicData = musicAPI.MiguAPI("周杰伦")
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
