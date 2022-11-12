package main

import (
	"fyneMusic/music"
	"log"
)

func init()  {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main()  {
	go music.MyPlayer.PlayMusic()           // 打开播放器
	go music.MyPlayer.InitPlayList()        // 异步加载数据
	go music.MyPlayer.RandomPlay()          // 随机播放
	go music.MyPlayer.UpdateProgressLabel() // 动态更新进度条、歌词
	music.RunApp()                          // 运行UI
}
