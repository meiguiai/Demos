package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	client "github.com/influxdata/influxdb1-client/v2"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Reader interface {
	Read(rc chan  []byte)
}

type Writer interface {
	Write(wc chan  *Message)
} 

type LogProcess struct {
	rc chan []byte
	wc chan *Message
	read Reader
	write Writer
}

type ReadFromFile struct {
	path string
}

type WriteToInfluxDB struct {
	influxDBDsn string  //influx data source
} 

type Message struct {
	timeLocal                       time.Time
	bytesSent					    int
	path, method, scheme, status 	string
	upstreamTime, requestTime		float64
}
// 系统状态监控
type SystemInfo struct {
	HandleLine   int     `json:"handleLine"`   // 总处理日志行数
	Tps          float64 `json:"tps"`          // 系统吞出量
	ReadChanLen  int     `json:"readChanLen"`  // read channel 长度
	WriteChanLen int     `json:"writeChanLen"` // write channel 长度
	RunTime      string  `json:"runTime"`      // 运行总时间
	ErrNum       int     `json:"errNum"`       // 错误数
}

const (
	TypeHandleLine = 0
	TypeErrNum     = 1
)

var TypeMonitorChan = make(chan int, 200)

type Monitor struct {
	startTime time.Time
	data SystemInfo
	tpsSli []int
}

func (m *Monitor) SystemInfoStart(lp *LogProcess) {
	go func() {
		for n := range TypeMonitorChan {
			switch n {
			case TypeErrNum:
				m.data.ErrNum += 1
			case TypeHandleLine:
				m.data.HandleLine += 1
			}
		}
	}()

	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for {
			<-ticker.C
			m.tpsSli = append(m.tpsSli, m.data.HandleLine)
			if len(m.tpsSli) > 2 {
				m.tpsSli = m.tpsSli[1:]
			}
		}
	}()

	http.HandleFunc("/monitor", func(writer http.ResponseWriter, request *http.Request) {
		m.data.RunTime = time.Now().Sub(m.startTime).String()
		m.data.ReadChanLen = len(lp.rc)
		m.data.WriteChanLen = len(lp.wc)

		if len(m.tpsSli) >= 2 {
			m.data.Tps = float64(m.tpsSli[1]-m.tpsSli[0]) / 5
		}

		ret, _ := json.MarshalIndent(m.data, "", "\t")
		_, err := io.WriteString(writer, string(ret))
		if err != nil {
			log.Fatal("io.WriteString",err)
		}
	})

 	err := http.ListenAndServe(":9193", nil)
	if err != nil {
		log.Fatal("ListenAndServe",err)
	}
}


func (w *WriteToInfluxDB) Write(wc chan  *Message) {
	// 写入模块
	infSli := strings.Split(w.influxDBDsn, "@")

	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     infSli[0],
		Username: infSli[1],
		Password: infSli[2],
	})
	if err != nil {
		log.Fatal(err)
	}

	for v := range wc {
		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  infSli[3],
			Precision: infSli[4],
		})
		if err != nil {
			log.Fatal(err)
		}

		// Create a point and add to batch
		// Tags: Path, Method, Scheme, Status
		tags := map[string]string{"Path": v.path, "Method": v.method, "Scheme": v.scheme, "Status": v.status}
		// Fields: UpstreamTime, RequestTime, BytesSent
		fields := map[string]interface{}{
			"UpstreamTime": v.upstreamTime,
			"RequestTime":  v.requestTime,
			"BytesSent":    v.bytesSent,
		}

		pt, err := client.NewPoint("nginx_log", tags, fields, v.timeLocal)
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		// Write the batch
		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}

		log.Println("write success!")
	}
}

func (r * ReadFromFile) Read(rc chan  []byte) {
	// 读取模块
	// 打开文件
	f,err := os.Open(r.path)
	if err != nil {
		panic(fmt.Sprintf("open file error:%s",err.Error()))
	}
	
	//从文件末尾开始读取文件内容
	//字符指针移动到末尾
	_,_ = f.Seek(0,2)
	rd := bufio.NewReader(f)

	for {
		line, err := rd.ReadBytes('\n')
		// 结尾错误判断
		if err == io.EOF {
			time.Sleep(500*time.Millisecond)
			continue
		} else if err != nil{
			panic(fmt.Sprintf("ReadBytes error:%s",err.Error()))
		}
		TypeMonitorChan <- TypeHandleLine
		rc <- line[:len(line) -1]
	}
}

func (l *LogProcess) Process() {
	// 解析模块
	/**
	172.0.0.12 - - [04/Mar/2018:13:49:52 +0000] http "GET /foo?query=t HTTP/1.0" 200 2133 "-" "KeepAliveClient" "-" 1.005 1.854
	*/

	r := regexp.MustCompile(`([\d\.]+)\s+([^ \[]+)\s+([^ \[]+)\s+\[([^\]]+)\]\s+([a-z]+)\s+\"([^"]+)\"\s+(\d{3})\s+(\d+)\s+\"([^"]+)\"\s+\"(.*?)\"\s+\"([\d\.-]+)\"\s+([\d\.-]+)\s+([\d\.-]+)`)

	loc, _ := time.LoadLocation("Asia/Shanghai")
	for v := range l.rc {
		ret := r.FindStringSubmatch(string(v))
		if len(ret) != 14 {
			TypeMonitorChan <- TypeErrNum
			log.Println("FindStringSubMatch fail:", string(v))
			continue
		}

		message := &Message{}
		t, err := time.ParseInLocation("2006-01-02 15:04:05 +0000", ret[4], loc)
		if err != nil {
			TypeMonitorChan <- TypeErrNum
			log.Println("ParseInLocation fail:", err.Error(), ret[4])
			continue
		}
		message.timeLocal = t

		byteSent, _ := strconv.Atoi(ret[8])
		message.bytesSent = byteSent

		// GET /foo?query=t HTTP/1.0
		reqSli := strings.Split(ret[6], " ")
		if len(reqSli) != 3 {
			TypeMonitorChan <- TypeErrNum
			log.Println("strings.Split fail", ret[6])
			continue
		}
		message.method = reqSli[0]

		u, err := url.Parse(reqSli[1])
		if err != nil {
			log.Println("url parse fail:", err)
			TypeMonitorChan <- TypeErrNum
			continue
		}
		message.path = u.Path

		message.scheme = ret[5]
		message.status = ret[7]

		upstreamTime, _ := strconv.ParseFloat(ret[12], 64)
		requestTime, _ := strconv.ParseFloat(ret[13], 64)
		message.upstreamTime = upstreamTime
		message.requestTime = requestTime

		l.wc <- message
	}
}

func main() {
	
	var path, influxDsn string
	flag.StringVar(&path,"path","/Users/ylw/Demos/log_process/access.log","read file path")
	flag.StringVar(&influxDsn,"influxDsn","http://127.0.0.1:8086@enter@enter@log_process@s","influx data source")
	flag.Parse()
	
	r := &ReadFromFile{
		path: path,
	}
	
	w := &WriteToInfluxDB{
		influxDBDsn: influxDsn,
	}
	
	lp := &LogProcess {
		rc: make(chan []byte, 200),
		wc: make(chan *Message,200),
		read: r,
		write: w,
	}
	
	go lp.read.Read(lp.rc)
	for i := 0;i<2; i++  {
		go lp.Process()
	}
	for i := 0;i<4; i++ {
		go lp.write.Write(lp.wc)
	}
	
	m := &Monitor{
		startTime: time.Now(),
		data: SystemInfo{},
	}
	
	m.SystemInfoStart(lp)
}
