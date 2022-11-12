package music

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"net/url"
	"os"
	"strings"
)

// MakeMyMenu 菜单组件
func MakeMyMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	saveMenuItem := fyne.NewMenuItem("设置", func() {
		cw := a.NewWindow("设置")
		cw.Resize(fyne.NewSize(600,400))
		cw.CenterOnScreen()

		savePath := widget.NewEntry()
		v1 := a.Preferences().String("savePath")
		if _,err := os.Stat(v1);err == nil {
			savePath.SetText(v1)
		}else {
			savePath.SetText(MyPlayer.DownloadPath)
		}
		os.MkdirAll(savePath.Text, 0755)

		miguServer := widget.NewEntry()
		v2 := a.Preferences().String("miguServer")
		if _,err := os.Stat(v2);err == nil {
			miguServer.SetText(v2)
		}else {
			miguServer.SetText(MyPlayer.MiguServer)
		}

		neteaseServer := widget.NewEntry()
		v3 := a.Preferences().String("neteaseServer")
		if _,err := os.Stat(v3);err == nil {
			neteaseServer.SetText(v3)
		}else {
			neteaseServer.SetText(MyPlayer.NeteaseServer)
		}

		form = &widget.Form{
			SubmitText: "确定",
			CancelText: "取消",
			Items: []*widget.FormItem{
				{Text: "下载路径", Widget: savePath, HintText: "歌曲保存路径"},
				{Text: "咪咕服务器", Widget: miguServer, HintText: "咪咕API服务器地址"},
				{Text: "网易云服务器", Widget: neteaseServer, HintText: "网易云API服务器地址"},
			},
			OnSubmit: func() {
				if _,err := os.Stat(savePath.Text);err == nil {
					a.Preferences().SetString("downloadPath", strings.TrimRight(savePath.Text, "\\"))
					MyPlayer.DownloadPath = a.Preferences().String("downloadPath")
					os.MkdirAll(MyPlayer.DownloadPath, 0755)
				}

				a.Preferences().SetString("miguServer", miguServer.Text)
				a.Preferences().SetString("neteaseServer", neteaseServer.Text)
				MyPlayer.MiguServer = miguServer.Text
				MyPlayer.NeteaseServer = neteaseServer.Text
				os.MkdirAll(savePath.Text, 0755)

				cw.Close()
			},
			OnCancel: func() {
				cw.Close()
			},
		}

		cw.SetContent(form)
		cw.Show()
	})

	helpMenuItem := fyne.NewMenuItem("github", func() {
		u, _ := url.Parse("https://github.com/ttmars/fyneMusic")
		_ = a.OpenURL(u)
	})

	// a quit item will be appended to our first (File) menu
	setting := fyne.NewMenu("菜单", saveMenuItem)
	help := fyne.NewMenu("帮助", helpMenuItem)
	mainMenu := fyne.NewMainMenu(
		setting,
		help,
	)
	return mainMenu
}
