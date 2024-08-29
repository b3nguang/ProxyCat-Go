package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/b3nguang/ProxyCat-Go/pkg/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/proxy"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var (
	proxies        []string
	proxyIndex     int
	rotateMode     string
	rotateInterval time.Duration
	mu             sync.Mutex
)

func init() {
	logger.InitLogger("INFO")
}

// LoadProxies 从指定路径加载代理列表
// filePath: 代理文件路径
func loadProxies(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Fatal("🏳 Error loading proxy file:", err)
	}
	defer file.Close()

	var proxies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxies = append(proxies, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Fatal("🏳 Error reading proxy file:", err)
	}
	return proxies
}

// RotateProxies 根据设定的时间间隔和模式轮换代理
// interval: 代理轮换的时间间隔
func rotateProxies(interval time.Duration) {
	for {
		time.Sleep(interval)
		mu.Lock()
		if rotateMode == "cycle" {
			proxyIndex = (proxyIndex + 1) % len(proxies)
		} else if rotateMode == "once" && proxyIndex < len(proxies)-1 {
			proxyIndex++
		}
		logger.Info("🔀 Switched to proxy:", proxies[proxyIndex])
		mu.Unlock()
	}
}

// GetCurrentProxy 获取当前使用的代理
func getCurrentProxy() string {
	mu.Lock()
	defer mu.Unlock()
	return proxies[proxyIndex]
}

// BuildCompleteURL 构建完整的请求URL
func buildCompleteURL(r *http.Request) string {
	if r.URL.IsAbs() {
		return r.URL.String()
	}
	return fmt.Sprintf("%s://%s%s", "http", r.Host, r.URL.RequestURI())
}

// ProxyMiddleware 处理HTTP代理请求的中间件
func proxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentProxy := getCurrentProxy()

		proxyURL, err := url.Parse(currentProxy)
		if err != nil {
			c.String(http.StatusInternalServerError, "🏳 Invalid proxy URL")
			logger.Error("🏳 Invalid proxy URL:", err)
			c.Abort()
			return
		}

		logger.Info("🙋‍ Handling request:", c.Request.Method, c.Request.URL.String(), "via proxy:", currentProxy)

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		client := &http.Client{Transport: transport}
		completeURL := buildCompleteURL(c.Request)
		req, err := http.NewRequest(c.Request.Method, completeURL, c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			logger.Error("🏳 Failed to create new request:", err)
			c.Abort()
			return
		}

		for name, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(name, value)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			logger.Error("🏳 Request failed:", err)
			c.Abort()
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			logger.Error("🏳 Failed to copy response body:", err)
			c.Abort()
			return
		}

		logger.Info("📡 Response:", c.Request.URL.String(), resp.StatusCode)
	}
}

// ConnectHandler 处理CONNECT请求
func connectHandler(c *gin.Context) {
	currentProxy := getCurrentProxy()
	proxyURL, err := url.Parse(currentProxy)
	if err != nil {
		c.String(http.StatusInternalServerError, "Invalid proxy URL")
		logger.Error("🏳 Invalid proxy URL:", err)
		return
	}

	logger.Info("Handling CONNECT request:", c.Request.Host, "via proxy:", currentProxy)

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		logger.Error("🏳 Failed to create dialer:", err)
		return
	}

	destConn, err := dialer.Dial("tcp", c.Request.Host)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		logger.Error("🏳 Failed to dial destination:", err)
		return
	}
	defer destConn.Close()

	hijacker, ok := c.Writer.(http.Hijacker)
	if !ok {
		c.String(http.StatusInternalServerError, "Webserver doesn't support hijacking")
		logger.Error("🏳 Webserver doesn't support hijacking")
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		logger.Error("🏳 Hijacking failed:", err)
		return
	}
	defer clientConn.Close()

	c.Status(http.StatusOK)

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)

	logger.Info("🎉 CONNECT established for:", c.Request.Host)
}

// Transfer 数据传输
func transfer(destination net.Conn, source net.Conn) {
	defer destination.Close()
	defer source.Close()
	_, err := io.Copy(destination, source)
	if err != nil {
		logger.Error("🏳 Error during data transfer:", err)
	}
}

func main() {
	port := flag.Int("p", 1080, "Listening port")
	mode := flag.String("m", "cycle", "Proxy rotation mode: cycle or once")
	interval := flag.Int("t", 60, "Proxy rotation interval (seconds)")
	flag.Parse()

	proxies = loadProxies("ip.txt")
	if len(proxies) == 0 {
		logger.Fatal("No proxies found")
	}

	rotateMode = *mode
	rotateInterval = time.Duration(*interval) * time.Second

	go rotateProxies(rotateInterval)

	// 创建 Gin Engine 并禁用 Gin 的默认日志中间件
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 添加自定义的日志中间件
	r.Use(proxyMiddleware())
	r.Any("/connect", connectHandler)

	logger.Info("🚀 Listening on port:", *port, "Proxy rotation mode:", *mode, "Proxy rotation interval:", *interval, "seconds")
	logger.Info("🤝 Initial proxy:", getCurrentProxy())

	r.Run(fmt.Sprintf(":%d", *port))
}
