package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/b3nguang/ProxyCat-Go/pkg/logger"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"golang.org/x/net/proxy"
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

// ProxyHandler 处理HTTP代理请求
func proxyHandler(w http.ResponseWriter, r *http.Request) {
	currentProxy := getCurrentProxy()

	proxyURL, err := url.Parse(currentProxy)
	if err != nil {
		http.Error(w, "🏳 Invalid proxy URL", http.StatusInternalServerError)
		logger.Error("🏳 Invalid proxy URL:", err)
		return
	}

	logger.Info("🙋‍ Handling request:", r.Method, r.URL.String(), "via proxy:", currentProxy)

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{Transport: transport}
	completeURL := buildCompleteURL(r)
	req, err := http.NewRequest(r.Method, completeURL, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("🏳 Failed to create new request:", err)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("🏳 Request failed:", err)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Set(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("🏳 Failed to copy response body:", err)
		return
	}

	logger.Info("📡 Response:", r.URL.String(), resp.StatusCode)
}

func connectHandler(w http.ResponseWriter, r *http.Request) {
	currentProxy := getCurrentProxy()
	proxyURL, err := url.Parse(currentProxy)
	if err != nil {
		http.Error(w, "Invalid proxy URL", http.StatusInternalServerError)
		logger.Error("🏳 Invalid proxy URL:", err)
		return
	}

	logger.Info("Handling CONNECT request:", r.Host, "via proxy:", currentProxy)

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("🏳 Failed to create dialer:", err)
		return
	}

	destConn, err := dialer.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("🏳 Failed to dial destination:", err)
		return
	}
	defer destConn.Close()

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Webserver doesn't support hijacking", http.StatusInternalServerError)
		logger.Error("🏳 Webserver doesn't support hijacking")
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logger.Error("🏳 Hijacking failed:", err)
		return
	}
	defer clientConn.Close()

	w.WriteHeader(http.StatusOK)

	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)

	logger.Info("🎉 CONNECT established for:", r.Host)
}

func transfer(destination net.Conn, source net.Conn) {
	defer destination.Close()
	defer source.Close()
	_, err := io.Copy(destination, source)
	if err != nil {
		logger.Error("🏳 Error during data transfer:", err)
	}
}

func startProxyServer(port int) {
	http.HandleFunc("/", proxyHandler)
	http.HandleFunc("/connect", connectHandler)

	address := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: address}

	logger.Info("😀 Starting proxy server on port:", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("🏳 Could not listen on", address, ":", err)
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

	logger.Info("🚀 Listening on port:", *port, "Proxy rotation mode:", *mode, "Proxy rotation interval:", *interval, "seconds")
	logger.Info("🤝 Initial proxy:", getCurrentProxy())

	startProxyServer(*port)
}
