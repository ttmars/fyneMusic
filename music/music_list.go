package music

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var ml *widget.List
func MakeMusicList() fyne.CanvasObject {
	ml = widget.NewList(
		func() int {
			return len(MyPlayer.PlayList)
		},
		func() fyne.CanvasObject {
			titleLabel := widget.NewLabelWithStyle("音乐标题", fyne.TextAlignLeading, fyne.TextStyle{Bold:true})
			singerLabel := widget.NewLabel("歌手")
			linkLabel := widget.NewHyperlink("播放", nil)
			albumLabel := widget.NewHyperlink("专辑", nil)
			downloadLabel := widget.NewHyperlink("标准", nil)
			flacDownloadLabel := widget.NewHyperlink("无损", nil)

			down := container.NewGridWithColumns(2, downloadLabel, flacDownloadLabel)
			musicLabel := container.NewGridWithColumns(2, linkLabel, down)
			musicLabel = container.NewGridWithColumns(4, titleLabel, singerLabel,albumLabel,musicLabel)
			return musicLabel
		},
		func(id widget.ListItemID, Item fyne.CanvasObject) {
			if id >= len(MyPlayer.PlayList) {
				return
			}
			d := MyPlayer.PlayList[id]
			Item.(*fyne.Container).Objects[0].(*widget.Label).SetText(d.Name)
			Item.(*fyne.Container).Objects[1].(*widget.Label).SetText(d.Singer)
			Item.(*fyne.Container).Objects[2].(*widget.Hyperlink).SetText(d.AlbumName)
			Item.(*fyne.Container).Objects[2].(*widget.Hyperlink).SetURLFromString(d.AlbumPic)
			c := Item.(*fyne.Container).Objects[3].(*fyne.Container)
			c.Objects[0].(*widget.Hyperlink).OnTapped = func() {
				MyPlayer.MusicChan <- d
			}
			cc := c.Objects[1].(*fyne.Container)
			cc.Objects[0].(*widget.Hyperlink).OnTapped = func() {
				path := MyPlayer.DownloadPath + "\\" + d.Name + "_" + d.Singer + ".mp3"
				err := DownloadMusic(d.Audio, path)
				if err != nil {
					dialog.ShowInformation("下载失败!", err.Error(), Window)
				}else{
					dialog.ShowInformation("下载成功!", path, Window)
				}
			}
			cc.Objects[1].(*widget.Hyperlink).OnTapped = func() {
				path := MyPlayer.DownloadPath + "\\" + d.Name + "_" + d.Singer + ".flac"
				err := DownloadMusic(d.Flac, path)
				if err != nil {
					dialog.ShowInformation("下载失败!", err.Error(), Window)
				}else{
					dialog.ShowInformation("下载成功!", path, Window)
				}
			}
			if d.Flac == "" {
				cc.Objects[1].(*widget.Hyperlink).Hide()
			}
		},
	)

	return ml
}

