package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
	"github.com/RyeZhu/sego"
	"runtime"
)

var (
	addr      = flag.String("addr", ":8080", "TCP address to listen to")
	dict      = flag.String("dict", "../data/dict.txt", "词典文件")
	compress  = flag.Bool("compress", false, "Whether to enable transparent response compression")
	segmenter = sego.Segmenter{}
)

type FastJsonResponse struct {
	Spam bool `json:"spam"`
}
type JsonResponse struct {
	Spam     bool       `json:"spam"`
	Segments []*Segment `json:"segments"`
}

type Segment struct {
	Text string `json:"text"`
	Pos  string `json:"pos"`
}

func main() {
	flag.Parse()

	// 将线程数设置为CPU数
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 初始化分词器
	segmenter.LoadDictionary(*dict)

	// 初始化 fasthttp request
	h := requestHandler
	if *compress {
		h = fasthttp.CompressHandler(h)
	}

	log.Print("服务器启动")

	if err := fasthttp.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

func requestHandler(ctx *fasthttp.RequestCtx) {

	text := string(ctx.Path())

	switch text {
	case "/favicon.ico":
		requestFavicon(ctx)
	default:
		requestSensitive(ctx)
	}
}

func requestFavicon(ctx *fasthttp.RequestCtx) {

}

func requestSensitive(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("application/json; charset=utf-8")
	// 得到要分词的文本

	text := string(ctx.Path())

	fmt.Fprintf(ctx, requestFilterSensitive(text))
}

func requestFilterSensitive(text string) string {
	spam := false

	size := len(text)
	if size > 200 {
		size = 200
	}

	// 分词
	segments := segmenter.Segment([]byte(text[1:size]), true)

	// 整理为输出格式
	ss := []*Segment{}
	for _, segment := range segments {
		if !spam && segment.Token().Pos() != "x" {
			spam = true
			break
		}
		ss = append(ss, &Segment{Text: segment.Token().Text(), Pos: segment.Token().Pos()})
	}
	//response, _ := json.Marshal(&JsonResponse{Spam: spam, Segments: ss})
	response, _ := json.Marshal(&FastJsonResponse{Spam: spam})
	return string(response)
}
