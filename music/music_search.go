package music

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"time"
)

var sw MusicSearch
type MusicSearch struct {
	SearchEntry *widget.Entry
	SearchEngine *widget.SelectEntry
	SearchSubmit *widget.Button
	SearchModCheck *widget.Check
}
func MakeMusicSearch()(c fyne.CanvasObject)  {
	sw.SearchEntry = widget.NewEntry()
	sw.SearchEntry.SetPlaceHolder("遇见萤火")

	sw.SearchEngine = widget.NewSelectEntry([]string{"网易云", "咪咕"})
	sw.SearchEngine.SetText("网易云")

	sw.SearchSubmit = widget.NewButtonWithIcon("搜索",theme.SearchIcon(), func() {
		searchFunc(sw.SearchEngine.Text, sw.SearchEntry.Text)
	})

	sw.SearchModCheck = widget.NewCheck("单曲循环", func(b bool) {
		if b {
			MyPlayer.PlayMode = 1
		}else {
			MyPlayer.PlayMode = 2
		}
	})
	c = container.NewBorder(nil,nil,nil,sw.SearchSubmit, sw.SearchEngine)
	c = container.NewGridWithColumns(5, sw.SearchEntry, c)
	c = container.NewBorder(nil,nil,nil, sw.SearchModCheck, c)

	titleLabel := widget.NewLabelWithStyle("音乐标题", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	singerLabel := widget.NewLabel("歌手")
	linkLabel := widget.NewLabel("点播")
	albumLabel := widget.NewLabel("专辑")
	downloadLabel := widget.NewLabelWithStyle("标准", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	flacDownloadLabel := widget.NewLabelWithStyle("无损", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
	down := container.NewGridWithColumns(2, downloadLabel, flacDownloadLabel)
	musicLabel := container.NewGridWithColumns(2, linkLabel, down)
	musicLabel = container.NewGridWithColumns(4, titleLabel, singerLabel,albumLabel,musicLabel)

	c = container.NewVBox(c, musicLabel)
	return
}

func searchFunc(eg, kw string)  {
	log.Println("搜索：", eg, kw)
	if kw == "" {
		return
	}
	MyPlayer.KeyWord = kw
	MyPlayer.SearchAPI = eg
	sw.SearchSubmit.Disable()
	defer sw.SearchSubmit.Enable()

	// 重新请求数据、创建组件并刷新
	if eg == "网易云" {
		if x, found := MyPlayer.SearchCache.Get("网易云"+kw); found {
			MyPlayer.PlayList = x.([]Song)
		}else{
			cur := time.Now()
			MyPlayer.PlayList = NeteaseAPI(kw)
			fmt.Println(len(MyPlayer.PlayList), kw)
			log.Println("网易云请求耗时：", time.Since(cur))
			if len(MyPlayer.PlayList) == 1 {
				dialog.ShowInformation("搜索失败", "网易云API服务器出错.", Window)
				return
			}else{
				MyPlayer.SearchCache.SetDefault("网易云"+kw, MyPlayer.PlayList)
			}
		}
	}else{
		if x, found := MyPlayer.SearchCache.Get("咪咕"+kw); found {
			MyPlayer.PlayList = x.([]Song)
		}else{
			cur := time.Now()
			MyPlayer.PlayList = MiguAPI(kw)
			log.Println("咪咕耗时：", time.Since(cur))
			if len(MyPlayer.PlayList) == 1 {
				dialog.ShowInformation("搜索失败", "咪咕API服务器出错.", Window)
				return
			}else{
				MyPlayer.SearchCache.SetDefault("咪咕"+kw, MyPlayer.PlayList)
			}
		}
	}

	ml.Refresh()

	//MyPlayer.DoneChan <- true		// 搜索后自动随机播放
}