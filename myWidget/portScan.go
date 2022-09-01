package myWidget

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"net"
	"strconv"
	"sync"
	"time"
)

var progress *widget.ProgressBar			// 进度条
var text *widget.Entry						// 消息框
var form *widget.Form

var wg sync.WaitGroup						// 同步子协程
var MaxCur = 1000							// 设置最大并发数，控制扫描速度
var ch = make(chan bool, MaxCur)

// MakePortScan 端口扫描界面
func MakePortScan(win fyne.Window) fyne.CanvasObject {
	host := widget.NewEntry()
	//host.SetPlaceHolder("baidu.com")
	host.SetText("")

	port := widget.NewEntry()
	port.SetText("1000")

	timeOut := widget.NewEntry()
	timeOut.SetText("3")

	progress = widget.NewProgressBar()
	progress.TextFormatter = func() string {
		return fmt.Sprintf("%d/%d", int(progress.Value), int(progress.Max))
	}
	progress.Min = 0
	progress.Max = 65535
	progress.Value = 0

	text = widget.NewMultiLineEntry()
	form = &widget.Form{
		SubmitText: "开始扫描",
		Items: []*widget.FormItem{
			{Text: "Host", Widget: host, HintText: "IP地址或域名"},
			{Text: "Port", Widget: port, HintText: "端口范围:0~Port,最大值:0~65535"},
			{Text: "TimeOut", Widget: timeOut, HintText: "TCP超时设置"},
		},
		OnSubmit: func() {
			form.Disable()
			progress.Value = 0
			progress.Refresh()
			p,err1 := strconv.Atoi(port.Text)
			t,err2 := strconv.Atoi(timeOut.Text)
			if err1 != nil || err2 != nil || host.Text == "" {
				form.Enable()
				return
			}
			portScan(host.Text, p, t)
			form.Enable()
		},
	}

	c := container.NewVBox(form, progress)
	cc := container.NewBorder(c,nil,nil,nil,text)

	return cc
}

func portScan(host string, MaxPort int, Timeout int)  {
	progress.Max = float64(MaxPort)
	var info string

	// 刷新文本框
	info += fmt.Sprintf("主机：%s 端口范围：0~%d 预计耗时：%ds\n", host, MaxPort, Timeout*MaxPort/MaxCur)
	text.SetText(info)
	text.Refresh()

	start := time.Now()
	for i:=0;i<=MaxPort;i++{
		wg.Add(1)
		ch <- true
		// 刷新进度条
		progress.Value = float64(i)
		progress.Refresh()
		go func(i int) {
			defer wg.Done()
			addr := fmt.Sprintf("%s:%d", host,i)
			c,err := net.DialTimeout("tcp", addr, time.Duration(Timeout) * time.Second)
			if err == nil {
				// 刷新文本框
				info += fmt.Sprintf("port %d is open!\n", i)
				text.SetText(info)
				text.Refresh()
				c.Close()
			}
			<-ch
		}(i)
	}
	wg.Wait()

	end := time.Since(start)
	// 刷新文本框
	info += fmt.Sprintf("扫描结束，耗时：%s\n", end)
	text.SetText(info)
	text.Refresh()
}

func refreshEntry(info string, entry *widget.Entry)  {
	entry.SetText(info)
	entry.Refresh()
}
