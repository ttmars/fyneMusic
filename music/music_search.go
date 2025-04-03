package music

import (
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
	SearchEntry  *widget.Entry
	SearchEngine *widget.SelectEntry
	SearchSubmit *widget.Button
}

func MakeMusicSearch() (c fyne.CanvasObject) {
	sw.SearchEntry = widget.NewEntry()
	sw.SearchEntry.SetPlaceHolder("孙露")

	sw.SearchEngine = widget.NewSelectEntry([]string{"网易云", "咪咕", "云盘"})
	sw.SearchEngine.SetText("网易云")

	sw.SearchSubmit = widget.NewButtonWithIcon("搜索", theme.SearchIcon(), func() {
		searchFunc(sw.SearchEngine.Text, sw.SearchEntry.Text)
	})

	c = container.NewBorder(nil, nil, nil, sw.SearchSubmit, sw.SearchEngine)
	c = container.NewGridWithColumns(5, sw.SearchEntry, c)

	titleLabel := widget.NewLabelWithStyle("标题", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	singerLabel := widget.NewLabel("歌手")
	albumLabel := widget.NewLabel("专辑")
	downloadLabel := widget.NewLabelWithStyle("下载", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	flacDownloadLabel := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	t1 := container.NewGridWithColumns(2, downloadLabel, flacDownloadLabel)
	t2 := container.NewGridWithColumns(3, titleLabel, singerLabel, albumLabel)
	t3 := container.NewBorder(nil, nil, nil, t1, t2)
	c = container.NewVBox(c, t3)
	return
}

func searchFunc(eg, kw string) {
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
		if x, found := MyPlayer.SearchCache.Get("网易云" + kw); found {
			MyPlayer.PlayList = x.([]Song)
		} else {
			cur := time.Now()
			MyPlayer.PlayList = NeteaseAPI(kw)
			log.Println("网易云请求耗时：", time.Since(cur))
			if len(MyPlayer.PlayList) == 0 {
				dialog.ShowInformation("搜索失败", "网易云API服务器出错.", Window)
				return
			} else {
				MyPlayer.SearchCache.SetDefault("网易云"+kw, MyPlayer.PlayList)
			}
		}
	} else if eg == "咪咕" {
		if x, found := MyPlayer.SearchCache.Get("咪咕" + kw); found {
			MyPlayer.PlayList = x.([]Song)
		} else {
			cur := time.Now()
			MyPlayer.PlayList = MiguAPI(kw)
			log.Println("咪咕耗时：", time.Since(cur))
			if len(MyPlayer.PlayList) == 0 {
				dialog.ShowInformation("搜索失败", "咪咕API服务器出错.", Window)
				return
			} else {
				MyPlayer.SearchCache.SetDefault("咪咕"+kw, MyPlayer.PlayList)
			}
		}
	} else {
		if x, found := MyPlayer.SearchCache.Get("云盘" + kw); found {
			MyPlayer.PlayList = x.([]Song)
		} else {
			cur := time.Now()
			MyPlayer.PlayList = CloudAPI(kw)
			log.Println("云盘耗时：", time.Since(cur))
			if len(MyPlayer.PlayList) == 0 {
				dialog.ShowInformation("搜索失败", "云盘搜索出错.", Window)
				return
			} else {
				MyPlayer.SearchCache.SetDefault("云盘"+kw, MyPlayer.PlayList)
			}
		}
	}

	ml.Refresh()
	MyPlayer.MusicChan <- MyPlayer.PlayList[0]
	MyPlayer.CurrentSongIndex = 0
	ml.Select(0)
}
