package main

import (
	"bufio"
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var database Database

func readUrlsFromStdin(urls chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		urls <- scanner.Text()
	}
	close(urls)
}

func main() {
	err := database.Create()
	if err != nil {
		log.Fatal(err)
	}

	urls := make(chan (string), 10000)
	go readUrlsFromStdin(urls)

	setLimits()
	wg := &sync.WaitGroup{}
	databaseWg := &sync.WaitGroup{}

	go database.ReceiveApiResponses(databaseWg)

	for n := 0; n < 1024; n++ {
		wg.Add(1)

		go func(workerN int) {
			for url := range urls {
				var ipAddr net.IP

				h := sha1.New()
				h.Write([]byte(url))

				bs := fmt.Sprintf("%x", h.Sum(nil))
				fmt.Println(bs)

				// This is really really bad code
				// One should specify start and end of ip range
				// and this software should choose randomly, but I am lazy and this works
				aua := "2a0f:5382:1312:8" + bs[0:3] + ":" + bs[3:7] + ":" + bs[7:11] + ":" + bs[11:15] + ":" + bs[15:19]
				fmt.Println(aua)
				ipAddr = net.ParseIP(aua)

				fmt.Println("Fetching '" + url + "' with IP '" + ipAddr.String() + "'")
			}

			wg.Done()
		}(n)
	}

	wg.Wait()

	database.Close()
	databaseWg.Wait()
}

func getHttpClient(localAddr net.IP, remoteAddr string) *http.Client {
	netDialer := &net.Dialer{
		Timeout: 12 * time.Second,
	}
	netDialer.LocalAddr = &net.TCPAddr{
		IP: localAddr,
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return netDialer.DialContext(ctx, "tcp", net.JoinHostPort(remoteAddr, "443"))
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if remoteAddr == "" {
		transport.DialContext = (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext
	} else {
		transport.DialContext = dialContext
	}
	return &http.Client{
		Transport: transport,
	}
}

func fetch(url string, httpClient *http.Client) bool {
	startTime := time.Now()
	httpRequest, _ := http.NewRequest("GET", url, nil)
	httpRequest.Header.Set("User-Agent", "fetcher/1.0.0")

	hasError := 0
	retryCounter := 0

	var httpResponse *http.Response
	var err error

	for retryCounter < 5 {
		httpResponse, err = httpClient.Do(httpRequest)
		if err == nil {
			break
		}
		retryCounter += 1
	}

	if err != nil {
		hasError = 1
		fmt.Println(err)
		return false
	}

	body, err := ioutil.ReadAll(httpResponse.Body)
	if httpResponse.Body != nil {
		httpResponse.Body.Close()
	}

	if err != nil {
		hasError = 1
	}

	apiResponse := &ApiResponse{
		Url:         url,
		Status:      httpResponse.StatusCode,
		ResponseRaw: string(body),
		Duration:    time.Since(startTime),
		Date:        startTime,
		HasError:    hasError,
	}

	database.ApiResponseChan <- apiResponse

	return true
}
