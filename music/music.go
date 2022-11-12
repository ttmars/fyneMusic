package music

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyneMusic/myTheme"
	"github.com/faiface/beep"
	"net/http"
	"os"
)
var streamer beep.StreamSeekCloser
var musicFormat beep.Format
var musicStreamer *beep.Resampler		// 控制播放速度
var ctrl *beep.Ctrl						// 控制暂停
//var flacDownloadButton *widget.Button	// 下载按钮
//var downloadButton *widget.Button		// 下载按钮
var line3 *fyne.Container
var Window fyne.Window
var App fyne.App


func CreateApp(myApp fyne.App, parent fyne.Window) fyne.CanvasObject {
	c1 := MakeMusicSearch()
	c2 := MakeMusicList()
	c3 := MakeMusicPlayer()
	c := container.NewBorder(c1,c3,nil,nil,c2)
	return c
}

func RunApp() {
	App = app.NewWithID("ccccd")				// 创建APP
	Window = App.NewWindow("网易云音乐")			// 创建窗口
	App.SetIcon(myTheme.ResourceLogoJpg)			    	// 设置logo
	App.Settings().SetTheme(&myTheme.MyTheme{})			// 设置APP主题，嵌入字体，解决乱码
	Window.Resize(fyne.NewSize(1200,800))			// 设置窗口大小
	Window.CenterOnScreen()								// 窗口居中显示
	Window.SetMaster()									// 设置为主窗口
	//myWindow.SetCloseIntercept(func() {myWindow.Hide()})	// 设置窗口托盘显示
	//if desk, ok := myApp.(desktop.App); ok {
	//	m := fyne.NewMenu("MyApp",
	//		fyne.NewMenuItem("Show", func() {
	//			myWindow.Show()
	//		}))
	//	desk.SetSystemTrayMenu(m)
	//}

	initPreferences()
	Window.SetMainMenu(MakeMyMenu(App, Window))
	Window.SetContent(CreateApp(App, Window))

	Window.ShowAndRun()
}

//初始化Preferences变量
func initPreferences()  {
	downloadPath :=  App.Preferences().String("savePath")
	miguServer := App.Preferences().String("miguServer")
	neteaseServer := App.Preferences().String("neteaseServer")
	if _,err := os.Stat(downloadPath); err == nil {
		MyPlayer.DownloadPath = downloadPath
	}
	if _,err := os.Stat(miguServer); err == nil {
		MyPlayer.MiguServer = miguServer
	}
	if _,err := os.Stat(neteaseServer); err == nil {
		MyPlayer.NeteaseServer = neteaseServer
	}
}

// 字符串长度裁剪
func cutString(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}else{
		return string(r[:n])
	}
}

// CreateImage 创建一个图片
func CreateImage(pic string) *canvas.Image {
	r,err := http.Get(pic)
	if err != nil {
		return canvas.NewImageFromResource(theme.FyneLogo())
	}
	defer r.Body.Close()
	image := canvas.NewImageFromReader(r.Body, "jpg")
	//image.FillMode = canvas.ImageFillOriginal
	return image
}
