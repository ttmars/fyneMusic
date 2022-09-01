package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"fyneMusic/myWidget"
	"net/url"
	"os"
	"strings"
)

func init()  {
	os.Setenv("FYNE_FONT", "./static/font/simkai.ttf")
}

func main() {
	RunApp()
}

func RunApp()  {
	// 创建app和窗口
	myApp := app.NewWithID("io.fyne.demo")
	myWindow := myApp.NewWindow("music")

	myApp.SetIcon(theme.FyneLogo())						// 设置APP图标
	myWindow.Resize(fyne.NewSize(1200,800))		// 设置窗口大小
	myWindow.CenterOnScreen()							// 窗口居中显示
	myWindow.SetMaster()								// 设置为主窗口
	myWindow.SetMainMenu(MakeMyMenu(myApp, myWindow))		// 创建窗口菜单
	myWindow.SetCloseIntercept(func() {myWindow.Hide()})	// 关闭窗口时，隐藏窗口，而不是退出
	if desk, ok := myApp.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				myWindow.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}

	split := myWidget.MakeNav(myApp, myWindow)		// 导航栏
	myWindow.SetContent(split)						// 设置内容
	myWindow.ShowAndRun()							// 运行
}

// MakeMyMenu 菜单组件
func MakeMyMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	saveMenuItem := fyne.NewMenuItem("下载路径", func() {
		e := widget.NewEntryWithData(myWidget.SavePath)
		tmp,_ := myWidget.SavePath.Get()
		dialog.NewForm("修改下载路径", "确认", "取消", []*widget.FormItem{widget.NewFormItem(">", e)}, func(b bool) {
			if !b || !IsDir(e.Text) {
				myWidget.SavePath.Set(tmp)
			}else {
				myWidget.SavePath.Set(strings.TrimRight(e.Text, "\\"))
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

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}