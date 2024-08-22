![ProxyCat-Go](https://socialify.git.ci/b3nguang/ProxyCat-Go/image?description=1&descriptionEditable=%E4%B8%80%E6%AC%BE%E5%9F%BA%E4%BA%8E%20Golang%20%E9%87%8D%E6%9E%84%E7%9A%84%E9%AB%98%E6%80%A7%E8%83%BD%E7%9A%84%E4%BB%A3%E7%90%86%E6%B1%A0%E4%B8%AD%E9%97%B4%E4%BB%B6&font=Inter&forks=1&issues=1&language=1&logo=https%3A%2F%2Favatars.githubusercontent.com%2Fu%2F121670274%3Fs%3D400%26u%3D686132087f2e2324958b610f905a1b388478295b%26v%3D4&name=1&owner=1&pattern=Circuit%20Board&pulls=1&stargazers=1&theme=Dark)

## ✈️ 一、工具特性

ProxyCat-Go 是一个高性能的代理池中间件，重构自 Python 实现，利用 Go 语言的高并发优势来提高代理池的性能和稳定性。它支持 HTTP/HTTPS 代理和 SOCKS5 代理，能够自动轮换代理，适用于需要大量网络请求的场景，如网络爬虫和渗透测试等。

- **高性能**：使用 Go 的并发处理能力，提高代理池的响应速度和吞吐量。
- **支持多种代理类型**：支持 HTTP/HTTPS 代理和 SOCKS5 代理。
- **代理轮换**：自动轮换代理，支持循环模式和一次性使用模式。
- **可配置**：可以通过配置文件指定代理列表和轮换策略。
- **跨平台支持**：支持 Linux、macOS 和 Windows 系统。

## 🚨 二、配置

1. 创建 `ip.txt` 文件，添加你的代理地址，每行一个代理，格式为 `protocol://host:port`，例如：

   ```
   http://127.0.0.1:8080
   socks5://127.0.0.1:1080
   ```

2. 运行代理池中间件时，你可以通过命令行参数来配置轮换模式和轮换间隔时间：

   ``` bash
   ./ProxyCat-Go -p 1080 -m cycle -t 60
   ```

   - `-p` 指定监听端口，默认值为 `1080`。
   - `-m` 指定代理轮换模式：`cycle` 表示循环使用，`once` 表示用完即止。
   - `-t` 指定代理更换时间（秒），默认值为 `60` 秒。

## 🐉 三、使用

ProxyCat-Go 启动后，将会在指定端口监听 HTTP/HTTPS 请求。你可以配置你的应用程序使用该端口作为代理服务器。

## 🖐 四、免责声明

1. 如果您下载、安装、使用、修改本工具及相关代码，即表明您信任本工具
2. 在使用本工具时造成对您自己或他人任何形式的损失和伤害，我们不承担任何责任
3. 如您在使用本工具的过程中存在任何非法行为，您需自行承担相应后果，我们将不承担任何法律及连带责任
4. 请您务必审慎阅读、充分理解各条款内容，特别是免除或者限制责任的条款，并选择接受或不接受
5. 除非您已阅读并接受本协议所有条款，否则您无权下载、安装或使用本工具
6. 您的下载、安装、使用等行为即视为您已阅读并同意上述协议的约束

## 🙏 五、参考项目

[https://github.com/honmashironeko/ProxyCat](https://github.com/honmashironeko/ProxyCat)

在这里向 `ProxyCat` 项目献上最诚挚的敬意。

