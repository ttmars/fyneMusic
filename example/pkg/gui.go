package musicplayer

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type AppGUI struct {
	baseDir           string              // 文件目录
	songs             []string            // 歌曲集合
	curSong           *MusicEntry         // 当前歌曲
	currentSongName   *widget.Label       // 当前曲目名称
	progress          *widget.ProgressBar // 播放进度
	consumedTime      *widget.Label       // 已用时间
	remainedTime      *widget.Label       // 剩余时间
	playBtn           *widget.Button      // 播放
	paused            bool                // 是否暂停标志
	nextBtn           *widget.Button      // 下一首
	preBtn            *widget.Button      // 上一首
	forwardBtn        *widget.Button      // 快进
	backwardBtn       *widget.Button      // 快退
	songIdx           int                 // 当前歌曲序号
	appDir            string              // 程序运行目录
	newSongFlag       bool                // 新的一首歌
	endUpdateProgress chan bool           // 停止更新进度条
}

func (appui *AppGUI) Run() {

	a := app.New()

	appui.newSongFlag = true
	appui.songIdx = 0
	re, _ := os.Executable()
	appui.appDir = filepath.Dir(re)
	fmt.Println("pwd:" + appui.appDir)
	appui.songs = make([]string, 0, 10)
	appui.endUpdateProgress = make(chan bool)
	appui.baseDir = "music_res"
	appui.currentSongName = widget.NewLabel("--")
	appui.progress = widget.NewProgressBar()
	appui.consumedTime = widget.NewLabel("0")
	appui.remainedTime = widget.NewLabel("0")
	appui.playBtn = widget.NewButton("Play", appui.PlaySong)
	appui.paused = true
	appui.nextBtn = widget.NewButton("Next", appui.NextSong)
	appui.preBtn = widget.NewButton("Prev", appui.PrevSong)
	appui.forwardBtn = widget.NewButton("Forward", nil)
	appui.backwardBtn = widget.NewButton("Backward", nil)
	appui.progress.Min = 0
	appui.progress.Max = 100
	appui.progress.SetValue(0)

	files, _ := ioutil.ReadDir(appui.appDir + "/" + appui.baseDir)
	for _, onefile := range files {
		if onefile.IsDir() {
			// do nothing
		} else {
			// 放入曲库
			postfix := path.Ext(onefile.Name())
			if postfix == ".mp3" {
				appui.songs = append(appui.songs, onefile.Name())
			}
		}
	}

	// 显示第一首歌的名字
	if len(appui.songs) != 0 {
		appui.currentSongName.SetText(appui.songs[0])
	}

	w := a.NewWindow("MP3播放器")
	w.SetTitle("MP3 Player")

	w.SetContent(fyne.NewContainerWithLayout(layout.NewGridLayout(1),
		appui.currentSongName,
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, appui.consumedTime, appui.remainedTime),
			appui.consumedTime,
			appui.remainedTime,
			appui.progress,
		),
		fyne.NewContainerWithLayout(layout.NewGridLayout(5),
			appui.preBtn,
			appui.backwardBtn,
			appui.playBtn,
			appui.forwardBtn,
			appui.nextBtn,
		),
	))

	appui.curSong = &MusicEntry{}
	if len(appui.songs) != 0 {
		appui.curSong.Source = appui.appDir + "/" + appui.baseDir + "/" + appui.songs[appui.songIdx]
	}

	w.ShowAndRun()
}

// hooks
func (appui *AppGUI) PlaySong() {
	if appui.newSongFlag {
		appui.newSongFlag = false
		appui.curSong.Open()
		appui.remainedTime.SetText(appui.curSong.Format.SampleRate.D(appui.curSong.Streamer.Len()).Round(time.Second).String())
		// 播放音乐
		go appui.curSong.Play()
		// 更新进度条
		go appui.UpdateProcess()
	}

	if appui.paused == true {
		appui.playBtn.SetText("Pause")
		appui.paused = false
		appui.curSong.paused <- false
	} else {
		appui.playBtn.SetText("Play")
		appui.paused = true
		appui.curSong.paused <- true
	}
}

func (appui *AppGUI) UpdateProcess() {
	appui.progress.Min = 0
	appui.progress.Max = float64(appui.curSong.Streamer.Len())
	for {
		select {
		case <-appui.endUpdateProgress:
			return
		case <-time.After(time.Second):
			appui.progress.SetValue(appui.curSong.progress)
			appui.consumedTime.SetText(appui.curSong.Format.SampleRate.D(appui.curSong.Streamer.Position()).Round(time.Second).String())
		}
	}
}

func (appui *AppGUI) NextSong() {
	appui.songIdx = appui.songIdx + 1
	if appui.songIdx >= len(appui.songs) {
		appui.songIdx = 0
	}
	appui.Reset()
}

func (appui *AppGUI) PrevSong() {
	appui.songIdx = appui.songIdx - 1
	if appui.songIdx < 0 {
		appui.songIdx = len(appui.songs) - 1
	}
	appui.Reset()
}

func (appui *AppGUI) Reset() {
	appui.currentSongName.SetText(appui.songs[appui.songIdx])
	appui.curSong.Source = appui.appDir + "/" + appui.baseDir + "/" + appui.songs[appui.songIdx]
	appui.paused = true
	appui.playBtn.SetText("Play")
	if !appui.newSongFlag {
		appui.curSong.Stop()
		appui.endUpdateProgress <- true
	}
	appui.newSongFlag = true
}
