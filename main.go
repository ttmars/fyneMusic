package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyneMusic/musicAPI"
	"fyneMusic/myTheme"
	"fyneMusic/myWidget"
	"fyneMusic/tool"
)

func init()  {
	myWidget.MusicData = musicAPI.NeteaseAPI("纯音乐")		// 初始化数据后再渲染界面
}

func main()  {
	myApp := app.NewWithID("io.fyne.demo")							// 创建APP
	myWindow := myApp.NewWindow("网易云音乐")						// 创建窗口
	if !tool.IsDir(myApp.Preferences().String("SongSavePath")) {	// APP参数检查
		myApp.Preferences().SetString("SongSavePath", myWidget.BasePath)
	}

	myApp.SetIcon(myTheme.ResourceLogoJpg)			    	// 设置logo
	myApp.Settings().SetTheme(&myTheme.MyTheme{})			// 设置APP主题，嵌入字体，解决乱码
	myWindow.Resize(fyne.NewSize(1200,800))			// 设置窗口大小
	myWindow.CenterOnScreen()								// 窗口居中显示
	myWindow.SetMaster()									// 设置为主窗口
	myWindow.SetCloseIntercept(func() {myWindow.Hide()})	// 设置窗口托盘显示
	if desk, ok := myApp.(desktop.App); ok {
		m := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				myWindow.Show()
			}))
		desk.SetSystemTrayMenu(m)
	}

	myWindow.SetMainMenu(myWidget.MakeMyMenu(myApp, myWindow))		// 创建菜单
	myWindow.SetContent(myWidget.MakeNav(myApp, myWindow))			// 创建导航

	go myWidget.RandomPlay()		// 随机播放线程
	go myWidget.PlayMusic()			// 播放线程

	myWindow.ShowAndRun()
}
