package music

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"math"
)

var mp MusicPlayer
type MusicPlayer struct {
	PlayerLabel *widget.Hyperlink
	PlayerLyric *widget.Label
	PlayerPause *widget.Button
	PlayerSpeedLabel *widget.Hyperlink
	PlayerStartTime *widget.Label
	PlayerProgress *widget.Slider
	PlayerEndTime *widget.Label
	PlayerNext *widget.Button
	PlayerSpeedLeft *widget.Label
	PlayerSpeedSlider *widget.Slider
	PlayerSpeedRight *widget.Label
}

func MakeMusicPlayer()(c fyne.CanvasObject)  {
	// 第一行
	mp.PlayerLabel = widget.NewHyperlink("", nil)
	mp.PlayerLabel.OnTapped = func() {
		if mp.PlayerLyric.Visible() {
			mp.PlayerLyric.Hide()
		}else{
			mp.PlayerLyric.Show()
			mp.PlayerLyric.Refresh()
		}
	}
	mp.PlayerLyric = widget.NewLabel("")
	mp.PlayerPause = widget.NewButtonWithIcon("暂  停", theme.MediaPauseIcon(), func() {
		if ctrl != nil {
			ctrl.Paused = !ctrl.Paused
			if ctrl.Paused {
				mp.PlayerPause.SetIcon(theme.MediaPlayIcon())
				mp.PlayerPause.SetText("播  放")
			}else{
				mp.PlayerPause.SetIcon(theme.MediaPauseIcon())
				mp.PlayerPause.SetText("暂  停")
			}
		}
	})

	// 第二行
	mp.PlayerSpeedLabel = widget.NewHyperlink("1.0倍速", nil)
	mp.PlayerSpeedLabel.OnTapped = func() {
		if line3.Visible() {
			line3.Hide()
		}else{
			line3.Show()
			line3.Refresh()
		}
	}
	mp.PlayerStartTime = widget.NewLabel("00:00")
	mp.PlayerProgress = widget.NewSlider(0,10)
	mp.PlayerProgress .SetValue(0)
	mp.PlayerProgress.Step = 1
	mp.PlayerProgress.OnChanged = func(f float64) {
		if streamer != nil {
			t := math.Abs(float64(streamer.Position())-f)
			if  t > 200000 {
				ctrl.Paused = true
				streamer.Seek(int(f))
				ctrl.Paused = false
			}
		}
	}
	mp.PlayerEndTime = widget.NewLabel("00:00")
	mp.PlayerNext = widget.NewButtonWithIcon("下一曲", theme.MediaSkipNextIcon(), func() {
		mp.PlayerNext.Disable()
		defer mp.PlayerNext.Enable()
		MyPlayer.DoneChan <- true
	})

	// 第三行
	mp.PlayerSpeedLeft = widget.NewLabel("0.5")
	mp.PlayerSpeedRight = widget.NewLabel("2.0")
	mp.PlayerSpeedSlider = widget.NewSlider(0.5,2.0)
	mp.PlayerSpeedSlider.SetValue(1)
	mp.PlayerSpeedSlider.Step = 0.1
	mp.PlayerSpeedSlider.OnChanged = func(f float64) {
		MyPlayer.Speed = f
		musicStreamer.SetRatio(f)
		mp.PlayerSpeedLabel.SetText(fmt.Sprintf("%.1f倍速", f))
		//speedSliderLeftLabel.SetText(fmt.Sprintf("%.1f", f))
	}


	line1 := container.NewBorder(nil,nil,mp.PlayerLabel,mp.PlayerPause, container.NewCenter(mp.PlayerLyric))
	line2 := container.NewBorder(nil,nil,mp.PlayerStartTime,mp.PlayerEndTime,mp.PlayerProgress)
	line2 = container.NewBorder(nil,nil,mp.PlayerSpeedLabel,mp.PlayerNext, line2)
	line3 = container.NewBorder(nil,nil,mp.PlayerSpeedLeft,mp.PlayerSpeedRight,mp.PlayerSpeedSlider)
	line3.Hide()

	c = container.NewVBox(line1, line2)
	c = container.NewBorder(nil,line3,nil,nil,c)
	return
}

