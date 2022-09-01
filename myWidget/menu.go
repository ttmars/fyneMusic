package myWidget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"fyneMusic/tool"
	"net/url"
	"strings"
)

// MakeMyMenu 菜单组件
func MakeMyMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	saveMenuItem := fyne.NewMenuItem("下载路径", func() {
		e := widget.NewEntryWithData(SavePath)
		tmp,_ := SavePath.Get()
		dialog.NewForm("修改下载路径", "确认", "取消", []*widget.FormItem{widget.NewFormItem(">", e)}, func(b bool) {
			if !b || !tool.IsDir(e.Text) {
				SavePath.Set(tmp)
			}else {
				SavePath.Set(strings.TrimRight(e.Text, "\\"))
			}
		}, w).Show()
	})

	helpMenuItem := fyne.NewMenuItem("开发文档", func() {
		u, _ := url.Parse("https://developer.fyne.io")
		_ = a.OpenURL(u)
	})

	// a quit item will be appended to our first (File) menu
	setting := fyne.NewMenu("设置", saveMenuItem)
	help := fyne.NewMenu("帮助", helpMenuItem)
	mainMenu := fyne.NewMainMenu(
		setting,
		help,
	)
	return mainMenu
}
