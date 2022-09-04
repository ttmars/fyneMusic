package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyneMusic/musicAPI"
	"fyneMusic/myTheme"
	"fyneMusic/myWidget"
	"fyneMusic/tool"
	"log"
)


func main()  {
	myApp := app.NewWithID("hello,world!")				// 创建APP
	myWindow := myApp.NewWindow("网易云音乐")			// 创建窗口
	initPreferences(myApp,myWindow)							// 初始化全局变量
	log.Println("migu server:", musicAPI.MiguServer)
	log.Println("net server:", musicAPI.NeteaseServer)

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

	go myWidget.RandomPlay()				// 随机播放线程
	go myWidget.PlayMusic()					// 播放线程
	go myWidget.UpdateProgressLabel()		// 播放进度更新线程

	myWindow.ShowAndRun()			// 事件循环
}

// 初始化Preferences变量
func initPreferences(a fyne.App, w fyne.Window)  {
	myWidget.W = w
	if !tool.IsDir(a.Preferences().String("savePath")) {
		a.Preferences().SetString("savePath", myWidget.BasePath)
	}
	if a.Preferences().String("miguServer") == "" {
		a.Preferences().SetString("miguServer", "39.101.203.25:3400")
	}
	if a.Preferences().String("neteaseServer") == "" {
		a.Preferences().SetString("neteaseServer", "neteaseapi.youthsweet.com")
	}
	musicAPI.MiguServer = a.Preferences().String("miguServer")
	musicAPI.NeteaseServer = a.Preferences().String("neteaseServer")
	myWidget.SavePath = a.Preferences().String("savePath")
}