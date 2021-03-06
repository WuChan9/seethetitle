package main

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	flag "github.com/spf13/pflag"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/simplifiedchinese"
)

var ipv4Net net.IPNet
var ports []int32
var threadnum uint8
var timeWait int

func myUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s -n 192.168.0.1/24 -p 80,443\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %-20s %s\n", "-n, --net", "扫描网段(default:127.0.0.1/32)")
	fmt.Fprintf(os.Stderr, "  %-20s %s\n", "-p, --port", "扫描端口(default:80)")
	fmt.Fprintf(os.Stderr, "  %-20s %s\n", "--timeout", "请求超时(default:3s)")
	fmt.Fprintf(os.Stderr, "  %-20s %s\n", "-t, --thread", "并发数")

}
func argsInit() {
	flag.Int32SliceVarP(&ports, "port", "p", []int32{80}, "扫描目标端口")
	_, defaultIpv4Net, _ := net.ParseCIDR("127.0.0.1/32")
	flag.IPNetVarP(&ipv4Net, "net", "n", *defaultIpv4Net, "扫描网段")
	flag.Uint8VarP(&threadnum, "thread", "t", 16, "扫描并发数")
	flag.IntVar(&timeWait, "timeout", 3, "请求超时")

	flag.Parse()
}

func hostTitleCrawl(ip string, client *http.Client, sem chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	var url string
	//
	for _, port := range ports {
		if port == 443 {
			url = "https://" + ip
		} else {
			url = "http://" + ip + ":" + strconv.Itoa(int(port))
		}
		resp, err := client.Get(url)
		if err != nil {
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			continue
		}
		_, charsetName, _ := charset.DetermineEncoding(body, "")
		// windows-1252 也可能是
		if charsetName == "gbk" {
			body, err = simplifiedchinese.GBK.NewDecoder().Bytes(body)
			if err != nil {
				resp.Body.Close()
				continue
			}
		}

		r := regexp.MustCompile("<title>(?P<title>[\\S ]+?)</title>")

		res := r.FindSubmatch(body)
		if resp.StatusCode == 200 {
			if len(res) == 2 {
				fmt.Printf("[+] 200 %s %q\n", url, res[1])
			}
		} else if resp.StatusCode == 302 {
			fmt.Printf("[-] 302 %s %s\n", url, resp.Header.Get("Location"))
		} else {
			fmt.Printf("[-] %d %s\n", resp.StatusCode, url)
		}
		resp.Body.Close()
	}
	<-sem
}

func main() {
	//替换默认的 Usage
	flag.Usage = myUsage
	argsInit()

	sem := make(chan int, threadnum) //线程请求
	wg := sync.WaitGroup{}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// client 支持并发重用
	client := &http.Client{
		Timeout:   time.Duration(timeWait) * time.Second,
		Transport: tr,
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	mask := binary.BigEndian.Uint32(ipv4Net.Mask)
	start := binary.BigEndian.Uint32(ipv4Net.IP)

	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)
	// loop through addresses as uint32

	for i := start; i <= finish; i++ {
		wg.Add(1)
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		sem <- 1
		go hostTitleCrawl(ip.String(), client, sem, &wg)
	}
	wg.Wait()
}
