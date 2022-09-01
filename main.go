package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"fyneMusic/myWidget"
	"os"
)

func init()  {
	os.Setenv("FYNE_FONT", "./static/font/simkai.ttf")		// 字体设置
}

func main()  {
	myApp := app.NewWithID("io.fyne.demo")			// 创建APP
	myWindow := myApp.NewWindow("music")			// 创建窗口

	myApp.SetIcon(theme.FyneLogo())						// 设置APP图标
	myWindow.Resize(fyne.NewSize(1200,800))		// 设置窗口大小
	myWindow.CenterOnScreen()							// 窗口居中显示
	myWindow.SetMaster()								// 设置为主窗口

	//myWindow.SetCloseIntercept(func() {myWindow.Hide()})	// 设置窗口托盘显示
	//if desk, ok := myApp.(desktop.App); ok {
	//	m := fyne.NewMenu("MyApp",
	//		fyne.NewMenuItem("Show", func() {
	//			myWindow.Show()
	//		}))
	//	desk.SetSystemTrayMenu(m)
	//}

	myWindow.SetMainMenu(myWidget.MakeMyMenu(myApp, myWindow))		// 创建菜单
	myWindow.SetContent(myWidget.MakeNav(myApp, myWindow))			// 创建导航

	go myWidget.RandomPlay()		// 随机播放线程
	go myWidget.PlayMusic()			// 播放线程

	myWindow.ShowAndRun()
}

