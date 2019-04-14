package main

import (
	"fmt"
	"flag"
	"os"
	"net/http"
	"io"
	"io/ioutil"
	"time"
	"context"
	"net/url"
	"log"
)

type responseInfo struct{
	status int 
	bytes int64
	duration time.Duration
	serverName string
}

type summaryInfo struct{
	requested int64
	responsed int64
}

func main() {
	fmt.Println("Hello from my app")
	requests := flag.Int64("n", 1, "Number of requests to perform") 
	concurrency := flag.Int64("c", 1, "Number of multiple requests to make at a time")
    timelimit := flag.Int("t", 500, "Seconds to max. to spend on benchmarking. This implies -n 50000" )      
    timeout := flag.Int64("s", 30, "Seconds to max. wait for each response")       

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timelimit) * time.Second)
	defer cancel()
	
	fmt.Println(requests, concurrency)

	flag.Parse()

	if flag.NArg() == 0 || *requests == 0 || *requests < *concurrency {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	benchmarkProcess(ctx, *concurrency, *requests, *timeout)	
}

func benchmarkProcess(ctx context.Context, concurrency int64, requests int64, timeout int64){
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
	defer cancel()


	select {
	case <-ctx.Done():
		fmt.Println("Process time out")
		return
	default:
	}
	c := make(chan responseInfo)
	link := flag.Arg(0) 

	
	summary := summaryInfo{}
	for i := int64(0); i < concurrency; i++ {
		summary.requested++
		go checkLink(ctx,link, c)
	}

	for response := range c {
		if summary.requested < requests {
			summary.requested++
			go checkLink(ctx, link, c)
		}
		summary.responsed++
		fmt.Println(response)
		if summary.responsed == summary.requested {
			break
		}
	}
}

func checkLink(ctx context.Context, link string, c chan responseInfo){
	
	select{
	case <- ctx.Done():
		fmt.Println(" Response time out")
		return
	default:
	}
	start := time.Now()
	res, err := http.Get(link)

	if err != nil {
		panic(err)
	}
	read, _ := io.Copy(ioutil.Discard, res.Body)
	
	u ,err := url.Parse(link)
	if err != nil {
		log.Fatal(err)
	}

	c <- responseInfo{
		status:res.StatusCode,
		bytes:read,
		duration:time.Now().Sub(start),
		serverName:u.Hostname(),
	}
}  