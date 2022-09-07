### 项目简介

fyneMusic是一个用go实现的简约风、个性化定制的音乐播放器。编译后为单个可执行程序，大小只有30M左右且跨平台.

前端UI通过[fyne](https://github.com/fyne-io/fyne)实现，数据来源使用[NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)，通过[beep](https://github.com/faiface/beep)实现音乐播放.

### 主要功能

- 音乐搜索，支持网易云、咪咕两种数据来源
- 动态设置网易云、咪咕API服务器
- 点播指定歌曲、播放、暂停、下一曲等
- 歌词动态刷新显示
- 动态刷新进度条，并支持快进、快退
- 播放模式：单曲循环、随机播放（默认）
- 支持两种音质下载：标准音质MP3格式、无损音质flac格式
- 支持倍速播放，可通过滑动条进行更细粒度的速度调节！

### 编译运行

1. 克隆项目

```shell
git clone https://github.com/ttmars/fyneMusic.git
```

2. 编译需要go语言环境，以及C编译器：[https://developer.fyne.io/started/](https://developer.fyne.io/started/)
3. 进入项目根目录，运行程序

```shell
go mod tidy
go run main.go
```

### 效果图

![image](https://raw.githubusercontent.com/ttmars/image/06d0d3f72f0bda46c8584687546830bbaa3e79f6/github/fyne.png)
