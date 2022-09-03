package myWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"fyneMusic/musicAPI"
	"fyneMusic/tool"
	"net/url"
	"strings"
)

// MakeMyMenu 菜单组件
func MakeMyMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	saveMenuItem := fyne.NewMenuItem("设置", func() {
		cw := a.NewWindow("设置")
		cw.Resize(fyne.NewSize(600,400))
		cw.CenterOnScreen()

		savePath := widget.NewEntry()
		savePath.SetText(a.Preferences().String("savePath"))

		miguServer := widget.NewEntry()
		miguServer.SetText(a.Preferences().String("miguServer"))

		neteaseServer := widget.NewEntry()
		neteaseServer.SetText(a.Preferences().String("neteaseServer"))

		form = &widget.Form{
			SubmitText: "确定",
			CancelText: "取消",
			Items: []*widget.FormItem{
				{Text: "下载路径", Widget: savePath, HintText: "歌曲保存路径"},
				{Text: "咪咕服务器", Widget: miguServer, HintText: "咪咕API服务器地址"},
				{Text: "网易云服务器", Widget: neteaseServer, HintText: "网易云API服务器地址"},
			},
			OnSubmit: func() {
				a.Preferences().SetString("miguServer", miguServer.Text)
				a.Preferences().SetString("neteaseServer", neteaseServer.Text)
				if tool.IsDir(savePath.Text) {
					a.Preferences().SetString("savePath", strings.TrimRight(savePath.Text, "\\"))
				}
				musicAPI.MiguServer = a.Preferences().String("miguServer")
				musicAPI.NeteaseServer = a.Preferences().String("neteaseServer")
				SavePath = a.Preferences().String("savePath")
				cw.Close()
			},
			OnCancel: func() {
				cw.Close()
			},
		}

		cw.SetContent(form)
		cw.Show()
	})

	helpMenuItem := fyne.NewMenuItem("开发文档", func() {
		u, _ := url.Parse("https://developer.fyne.io")
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
