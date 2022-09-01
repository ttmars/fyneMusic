package myWidget

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)


var musicLength = 30					// 歌曲名称长度
var pauseButton *widget.Button			// 暂停按钮，控制样式
var searchSubmit *widget.Button			// 搜索按钮，防止重复点击
var musicName = binding.NewString()		// 当前播放歌曲名称

var musicList *container.Scroll			// 歌曲列表，刷新数据
var cs []fyne.CanvasObject

var speedLabel *widget.Label			// 倍速标签，控制样式
var speedSlider *widget.Slider			// 倍速滑动条，控制样式
var musicStreamer *beep.Resampler		// 控制播放速度
var ctrl *beep.Ctrl						// 控制暂停

var musicCh = make(chan string, 1)		// 控制播放
var doneCh = make(chan bool, 1)			// 控制随机播放

// 歌曲信息
var musicData = musicAPI(API["网易云"], "纯音乐")
type AutoGenerated []struct {
	Author string `json:"author"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Pic    string `json:"pic"`
	Src    string `json:"src"`
}
var API = map[string]string{
	"网易云": "http://www.youthsweet.com/netease/api?keyword=",
	"咪咕音乐":"http://www.youthsweet.com/migu/api?keyword=",
	"QQ音乐": "http://www.youthsweet.com/netease/api?keyword=",
}

// 下载路径
var BasePath,_ = filepath.Abs(".")
var SavePath = binding.NewString()
func init()  {
	SavePath.Set(BasePath)
}

var W fyne.Window
// RandomPlay 随机播放线程
func RandomPlay()  {
	doneCh <- true		// 启动后，随机播放歌曲
	for {
		select {
		case <-doneCh:
			rand.Seed(time.Now().Unix())
			if len(musicData) == 0 {
				dialog.ShowInformation("搜索失败!", "请更换关键词！", W)
				break
			}
			randomNum := rand.Intn(len(musicData))
			musicCh <- musicData[randomNum].Src
			musicName.Set("随机播放：" + musicData[randomNum].Name + "_" + musicData[randomNum].Author + ".mp3")
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

// MusicTable 歌曲列表组件
func MusicTable(parent fyne.Window) fyne.CanvasObject {
	W = parent
	// 搜索组件
	search := searchWidget(parent)

	// 音乐标签
	titleLabel := widget.NewLabelWithStyle("音乐标题", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	singerLabel := widget.NewLabelWithStyle("歌手", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	linkLabel := widget.NewLabelWithStyle("点播", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	downloadLabel := widget.NewLabelWithStyle("下载", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	musicLabel := container.NewGridWithColumns(2, linkLabel, downloadLabel)
	musicLabel = container.NewGridWithColumns(3, titleLabel, singerLabel, musicLabel)

	// 音乐列表
	for _,v := range musicData {
		t := oneMusic(v.Name, v.Author, v.Src, parent)
		cs = append(cs, t)
	}
	musicList = container.NewVScroll(container.NewVBox(cs...))		// 先按行布局，然后滚动

	// 播放器
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

	// 组合布局！
	c := container.NewVBox(search, musicLabel)
	cc := container.NewBorder(c, player, nil, nil, musicList)
	return cc
}

// 搜索组件
func searchWidget(parent fyne.Window)fyne.CanvasObject  {
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("纯音乐")

	//searchEngine := widget.NewSelectEntry([]string{"网易云", "咪咕音乐", "QQ音乐"})
	searchEngine := widget.NewSelectEntry([]string{"网易云", "咪咕音乐"})
	searchEngine.SetText("网易云")

	searchSubmit = widget.NewButtonWithIcon("搜索",theme.SearchIcon(), func() {
		searchSubmit.Disable()
		// 清空原有数据，重新渲染
		musicData = musicAPI(API[searchEngine.Text], searchEntry.Text)
		cs = cs[0:0]
		for _,v := range musicData {
			t := oneMusic(v.Name, v.Author, v.Src, parent)
			cs = append(cs, t)
		}
		musicList.Refresh()
		doneCh <- true		// 搜索后自动随机播放
		searchSubmit.Enable()
	})

	c := container.NewBorder(nil,nil,nil,searchSubmit, searchEngine)
	c = container.NewGridWithColumns(5, searchEntry, c)
	return c
}

// 单行歌曲组件
func oneMusic(v1, v2, v3 string, parent fyne.Window) fyne.CanvasObject {
	r := []rune(v1)
	if len(r) > musicLength {
		r = r[:musicLength]
		v1 = string(r)
	}
	titleLable := widget.NewLabel(v1)
	singerLable := widget.NewLabel(v2)
	linkLable := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		pauseButton.SetIcon(theme.MediaPauseIcon())
		pauseButton.SetText("暂停")
		musicName.Set("当前点播：" + v1 + "_" + v2 + ".mp3")
		speedLabel.SetText("1.0倍速")
		speedSlider.SetValue(1)
		musicCh <- v3
	})
	downloadLable := widget.NewButtonWithIcon("", theme.DownloadIcon(), func() {
		path,_ := SavePath.Get()
		path = path + "\\" + v1 + "_" + v2 + ".mp3"
		err := downloadMusic(v3, path)
		if err != nil {
			dialog.ShowInformation("下载失败!", err.Error(), parent)
		}else{
			dialog.ShowInformation("下载成功!", path, parent)
		}
	})


	c := container.NewGridWithColumns(2, linkLable, downloadLable)
	c = container.NewGridWithColumns(3, titleLable, singerLable,c)
	return c
}

// 下载歌曲
func downloadMusic(url, path string) error  {
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

// 网易云API
func musicAPI(u, key string)  AutoGenerated {
	if key == "" {
		return nil
	}
	var v AutoGenerated
	r,err := http.Get(u+key)
	if err != nil {
		log.Println(err)
		return v
	}
	defer r.Body.Close()
	b,err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return v
	}
	err = json.Unmarshal(b, &v)
	if err != nil {
		log.Println(err)
		return v
	}
	return v
}