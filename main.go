package main

import (
	"fmt"
	"strings"
	"github.com/valyala/fasthttp"
)

// 需要移除的隐私相关请求头
var privacyHeaders = []string{
	"X-Forwarded-For", "X-Real-IP", "CF-Connecting-IP", "CF-IPCountry",
	"CF-Worker", "True-Client-IP", "Forwarded", "Via", "X-Cluster-Client-IP",
	"X-Forwarded-Host", "X-Forwarded-Proto", "X-Originating-IP", "X-Remote-IP",
	"X-Remote-Addr", "X-Envoy-External-Address", "X-Amzn-Trace-Id",
	"X-Request-Id", "X-Correlation-Id",
}

func proxyHandler(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())[1:] // 去除开头的 '/'
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Invalid URL. Please use format: /https://api.example.com/path")
		return
	}

	// 代理请求
	client := &fasthttp.Client{}
	url := path
	ctx.Request.Header.Del("Host")

	// 移除隐私相关请求头
	for _, h := range privacyHeaders {
		ctx.Request.Header.Del(h)
	}

	// 发送代理请求
	resp := fasthttp.AcquireResponse()
	err := client.Do(&ctx.Request, resp)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadGateway)
		ctx.SetBodyString("Proxy request failed")
		return
	}

	// 复制响应
	ctx.SetStatusCode(resp.StatusCode())
	ctx.SetBodyRaw(resp.Body())

	// 移除隐私相关响应头
	for _, h := range privacyHeaders {
		resp.Header.Del(h)
	}
	resp.Header.VisitAll(func(k, v []byte) {
		ctx.Response.Header.SetBytesKV(k, v)
	})

	fasthttp.ReleaseResponse(resp)
}

func main() {
	port := ":3000"
	fmt.Println("Server is running on port" + port)
	fmt.Println("Example usage: http://localhost:3000/https://api.example.com/path")
	if err := fasthttp.ListenAndServe(port, proxyHandler); err != nil {
		panic(err)
	}
}
