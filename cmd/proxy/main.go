package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "9090"
	}
	target := os.Getenv("TARGET_URL")
	if target == "" {
		target = "https://ark.cn-beijing.volces.com/api/coding/v3"
	}

	// 打印带颜色的JSON
	pretty := func(prefix string, data []byte) {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err == nil {
			fmt.Printf("\n%s\n%s\n%s\n", strings.Repeat("─", 60), prefix, buf.String())
		} else {
			fmt.Printf("\n%s\n%s\n%s\n", strings.Repeat("─", 60), prefix, string(data))
		}
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		// 构造目标URL
		targetURL := target + strings.TrimPrefix(r.URL.Path, "")
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		// 读取请求体
		reqBody, _ := io.ReadAll(r.Body)
		r.Body.Close()

		// 打印请求
		fmt.Printf("\n%s\n", strings.Repeat("═", 80))
		fmt.Printf(">>> REQUEST  %s %s\n", r.Method, r.URL.String())
		fmt.Printf(">>> TARGET   %s %s\n", r.Method, targetURL)
		fmt.Println(">>> HEADERS:")
		for k, v := range r.Header {
			if k == "Authorization" {
				fmt.Printf("    %s: Bearer ***\n", k)
			} else {
				fmt.Printf("    %s: %s\n", k, strings.Join(v, ","))
			}
		}
		pretty(">>> BODY:", reqBody)

		// 转发请求
		proxyReq, _ := http.NewRequest(r.Method, targetURL, bytes.NewReader(reqBody))
		for k, v := range r.Header {
			proxyReq.Header[k] = v
		}

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			fmt.Printf(">>> ERROR: %v\n", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 读取响应
		respBody, _ := io.ReadAll(resp.Body)

		// 打印响应
		fmt.Println("<<< STATUS: ", resp.StatusCode)
		fmt.Println("<<< HEADERS:")
		for k, v := range resp.Header {
			fmt.Printf("    %s: %s\n", k, strings.Join(v, ","))
		}
		if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
			fmt.Printf("<<< BODY: [SSE stream - %d bytes]\n", len(respBody))
		} else {
			pretty("<<< BODY:", respBody)
		}
		fmt.Printf("%s\n", strings.Repeat("═", 80))

		// 返回响应
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
	}

	http.HandleFunc("/", handler)
	log.Printf("抓包代理启动: http://localhost:%s -> %s", port, target)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
