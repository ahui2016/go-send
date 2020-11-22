# go-send

快捷备忘 · 文件互传


## 主要功能

- 随手保存文件或简短备忘
- 方便个人设备间或熟人之间互传文件
- 支持 iPhone Shortcut (快捷指令)
- 另提供 go-send-cli 命令行工具，服务器、终端环境也可使用
- 支持快捷键发送系统剪贴板内容 (可发送纯文本或文件)


## demo 演示

- https://send.ai42.xyz
- 密码: abc
- 演示版会自动压缩图片，正式版则是上传原图
- 演示版限制单个文件 512KB 以下，数据库总容量 10MB, 正式版可自由设定
- 演示版限制文件 24 小时变灰，30 天自动删除，正式版可自由设定


## iPhone Shortcut (快捷指令)

- 接收最近一条文本消息
  - 需要用到的网址 `https://send.ai42.xyz/api/last-text?password=abc`
  - 具体设置方法如图 https://send.ai42.xyz/public/go-send-get-last.jpeg
  - 我设置的是接收消息后保存到剪贴板，但根据需要也可以很容易改成保存到记事本、提醒、日历等

- 发送文本消息
  - 需要用到的网址 `https://send.ai42.xyz/api/add-text`
  - 具体设置方法如图 https://send.ai42.xyz/public/go-send-add-text.jpeg
  - 如图所示，我设置的是发送剪贴板的内容

- 发送图片
  - 需要用到的网址 `https://send.ai42.xyz/api/add-photo`
  - 具体设置方法如图 https://send.ai42.xyz/public/go-send-add-photo.jpeg
  - 如图所示，我增加了一个缩小图片的流程，那是因为演示版只能接收很小的文件，正式版可发原图
  - 设置成功后，打开相册选择图片，点击发送，选择快捷指令即可
  - 目前每次只能发送一张图片，需要一次发送多张，或者需要发送视频、文档请使用网页版

- 注意，以上是使用了我的 demo 服务器，只是方便试用体验，真正使用还是需要用户自行搭建服务器的。


## 技术特点

- 后端采用 Go, 前端采用 jQuery 和 Bootstrap, 都是非常简单直白的代码
- 程序很小，不需要配置数据库，占用资源（内存、CPU）很少
- 前后端分离，后端 api 接口一律返回 json 给前端，一共只有 3 个 html 页面
- 由于是个简单的小项目，因此容易定制，有什么不满意的地方，用户可自行定制修改
- 没有 webpack, 没有 node_modules, 没有 Vue, React, 不是抗拒新技术，主要是因为没有复杂的功能，因此不过度使用这些专为复杂网页而设的技术。由于前后端分离，如果用户喜欢，也可以自行用 Vue, React 重写前端，后端完全不用修改。
- 代码里，凡是需要注意的地方都加了注释，如有疑问也欢迎问我


## Requirements

- 需要有自己的服务器
- 需要有自己的域名 (用于配置https)
- 安装 Go SDK (https://golang.org/doc/install)


## 安装运行

```sh
$ cd ~
$ git clone https://github.com/ahui2016/go-send.git
$ cd go-send && go build
$ ./go-send &
```

### 设置密码和端口

- 默认密码是 abc, 默认端口是 127.0.0.1:80
- 第一次运行 go-send 时，会在 $HOME 目录自动新建一个文件夹 gosend_data_folder, 并且在该文件夹内生成 config 文件，直接用文本编辑器修改 config 文件即可设置密码和端口，修改保存后，重启 go-send 生效。
  ```sh
  $ cd vim ~/gosend_data_folder/config (修改、保存、退出)
  $ killall go-send
  $ cd ~/go-send
  $ ./go-send &
  ```

### 设置 Nginx 及 https

- 本软件需要在浏览器里生成 SHA256, 而浏览器要求在 https 模式下才能使用 SHA256 的功能，因此必须配置 https
- 在 Nginx 配置文件 (通常是 /etc/nginx/nginx.conf) 中添加以下内容：
  ```
  server {
      listen 80;
      server_name your.domain.com;
      location / {
          proxy_pass http://127.0.0.1:80/;
      }
  }
  ```
- 使 nginx 的修改生效
  ```sh
  $ sudo nginx -s reload
  ```
- 如果已经安装 Certbot, 执行以下命令即可自动配置 https
  ```sh
  sudo certbot --nginx
  ```
- 如果未安装 Certbot, 看这里 https://certbot.eff.org 或者用其他方法配置 https


## go-send-cli 命令行

- 另外有一个配套的 go-send-cli 工具，属于额外功能，不安装也不影响 go-send 的正常使用
- 安装 go-send-cli 后，无需打开浏览器，在终端即可接收或发布文本消息或文件
- https://github.com/ahui2016/go-send-cli


## 发送系统剪贴板内容

- Windows
  - 详情请看 https://github.com/ahui2016/gosend.ahk
- iPhone/iPad
  - 可通过快捷命令实现该功能，具体方法见上文 “iPhone Shortcut (快捷指令)” 章节。
- Mac
  - 我没有 Mac, 因此没做 Mac 版，但我猜 Mac 里也有类似 AutoHotkey 或快捷指令的工具，
    Mac 用户可自行实现，应该不难。
  - 如果有人实现了 Mac 版，请反馈给我，在此先谢谢了！
  - 如果实在没人做，我也可能自己做（呃）。
  

## Go-send Project

由于 go-send 对 Windows, iOS, Android, Linux, macOS, Chrome, Firefox 等不同平台提供（或计划提供）
便捷的使用方法，因此很多功能不方便直接写在主程序主仓库里，显得稍有点乱，因此建了一个 project 页面汇总信息：

https://github.com/users/ahui2016/projects/1
