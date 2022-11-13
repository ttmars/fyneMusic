package music

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"net/http"
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
			albumLabel := widget.NewHyperlink("专辑", nil)
			downloadLabel := widget.NewHyperlink("标准", nil)
			flacDownloadLabel := widget.NewHyperlink("无损", nil)

			t1 := container.NewGridWithColumns(2, downloadLabel, flacDownloadLabel)
			t2 := container.NewGridWithColumns(3, titleLabel,singerLabel,albumLabel)
			c := container.NewBorder(nil,nil,nil,t1,t2)
			return c
		},
		func(id widget.ListItemID, Item fyne.CanvasObject) {
			if id >= len(MyPlayer.PlayList) {
				return
			}
			d := MyPlayer.PlayList[id]

			c2 := Item.(*fyne.Container).Objects[0].(*fyne.Container)
			c1 := Item.(*fyne.Container).Objects[1].(*fyne.Container)
			//fmt.Println(len(c1.Objects), len(c2.Objects))

			s := []rune(d.Name)
			if len(s) >=29 {
				s = s[:29]
			}
			c2.Objects[0].(*widget.Label).SetText(string(s))
			c2.Objects[1].(*widget.Label).SetText(d.Singer)
			c2.Objects[2].(*widget.Hyperlink).SetText(d.AlbumName)
			c2.Objects[2].(*widget.Hyperlink).OnTapped = func() {
				w := App.NewWindow("image")
				w.CenterOnScreen()
				w.SetContent(createImage(d.AlbumPic))
				w.Resize(fyne.NewSize(400,400))
				w.Show()
			}

			c1.Objects[0].(*widget.Hyperlink).OnTapped = func() {
				path := MyPlayer.DownloadPath + "\\" + d.Name + ".mp3"
				err := DownloadMusic(d.Audio, path)
				if err != nil {
					dialog.ShowInformation("下载失败!", err.Error(), Window)
				}else{
					dialog.ShowInformation("下载成功!", path, Window)
				}
			}
			c1.Objects[1].(*widget.Hyperlink).OnTapped = func() {
				path := MyPlayer.DownloadPath + "\\" + d.Name + ".flac"
				err := DownloadMusic(d.Flac, path)
				if err != nil {
					dialog.ShowInformation("下载失败!", err.Error(), Window)
				}else{
					dialog.ShowInformation("下载成功!", path, Window)
				}
			}
			if d.Flac == "" {
				c1.Objects[1].(*widget.Hyperlink).Hide()
			}
		},
	)

	ml.OnSelected = func(id widget.ListItemID) {
		MyPlayer.MusicChan <- MyPlayer.PlayList[id]
		MyPlayer.CurrentSongIndex = id
	}

	return ml
}

func createImage(pic string) *canvas.Image {
	r,err := http.Get(pic)
	if err != nil {
		return canvas.NewImageFromResource(theme.FyneLogo())
	}
	defer r.Body.Close()
	image := canvas.NewImageFromReader(r.Body, "jpg")
	//image.FillMode = canvas.ImageFillOriginal
	return image
}

