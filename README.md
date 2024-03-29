# go-send

快捷备忘 · 文件互传


## 更新 (2020-12-20)

- 后端从 net/http 改成 fiber, 运行效率更高，速度更快


## 主要功能

- 随手保存文件或简短备忘
- 方便个人设备间或熟人之间互传文件
- 支持 iPhone Shortcut (快捷指令)
- 另提供 go-send-cli 命令行工具，服务器、终端环境也可使用
- 支持快捷键发送系统剪贴板内容 (可发送纯文本或文件)


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
  $ vim ~/gosend_data_folder/config (修改、保存、退出)
  $ cd ~/go-send
  $ killall go-send && ./go-send &
  ```

### 设置 Nginx 及 https

- 本软件需要在浏览器里生成 SHA256, 而浏览器要求在 https 模式下才能使用 SHA256 的功能，因此必须配置 https
- 如果不使用发送文件的功能，只使用文本备忘功能，不配置 https 也行。
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

