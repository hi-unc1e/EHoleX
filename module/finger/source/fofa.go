package source

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
)

type Config struct {
	Email        string
	Fofa_token   string
	Fofa_timeout string
	ZoomEye_key  string
}

type AutoGenerated struct {
	Mode    string     `json:"mode"`
	Error   bool       `json:"error"`
	Query   string     `json:"query"`
	Page    int        `json:"page"`
	Size    int        `json:"size"`
	Results [][]string `json:"results"`
}

// 获取当前执行程序所在的绝对路径
func GetCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func GetConfig() Config {
	//创建一个空的结构体,将本地文件读取的信息放入
	c := &Config{}
	//创建一个结构体变量的反射
	cr := reflect.ValueOf(c).Elem()
	//打开文件io流
	f, err := os.Open(GetCurrentAbPathByExecutable() + "/config.ini")
	if err != nil {
		//log.Fatal(err)
		color.RGBStyleFromString("237,64,35").Println("[Error] Fofa configuration file error!!!")
		os.Exit(1)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	//我们要逐行读取文件内容
	s := bufio.NewScanner(f)
	for s.Scan() {
		//以=分割,前面为key,后面为value
		var str = s.Text()
		var index = strings.Index(str, "=")
		var key = str[0:index]
		var value = str[index+1:]
		//通过反射将字段设置进去
		cr.FieldByName(key).Set(reflect.ValueOf(value))
	}
	err = s.Err()
	if err != nil {
		log.Fatal(err)
	}
	//返回Config结构体变量
	return *c
}

func fofa_api(keyword string, email string, key string, page int, size int) string {
	input := []byte(keyword)
	encodeString := base64.StdEncoding.EncodeToString(input)
	api_request := fmt.Sprintf("https://fofa.info/api/v1/search/all?email=%s&page=%d&size=%d&key=%s&qbase64=%s&fields=ip,host,title,port,protocol", strings.Trim(email, " "), page, size, strings.Trim(key, " "), encodeString)
	return api_request
}

func fofahttp(url string, timeout string) *AutoGenerated {
	var itime, err = strconv.Atoi(timeout)
	if err != nil {
		log.Println("fofa超时参数错误: ", err)
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{
		Timeout:   time.Duration(itime) * time.Second,
		Transport: transport,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "*/*;q=0.8")
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2227.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	res := &AutoGenerated{}
	json.Unmarshal(result, &res)
	return res
}

func Fofaip(ips string) (urls []string) {
	color.RGBStyleFromString("244,211,49").Println("请耐心等待fofa搜索......")
	fofa := GetConfig()
	keyword := `ip="` + ips + `"`
	url := fofa_api(keyword, fofa.Email, fofa.Fofa_token, 1, 5)
	res := fofahttp(url, fofa.Fofa_timeout)
	//fmt.Println(url)
	if res.Size > 10000 {
		words := []string{" && protocol=\"http\"", " && protocol=\"https\"", " && protocol=\"unknown\""}
		for _, k := range words {
			keyword1 := keyword + k
			for i := 1; i <= 20; i++ {
				url := fofa_api(keyword1, fofa.Email, fofa.Fofa_token, i, 500)
				res := fofahttp(url, fofa.Fofa_timeout)
				if len(res.Results) > 0 {
					for _, value := range res.Results {
						if strings.Contains(value[1], "http") {
							urls = append(urls, value[1])
						} else {
							if k == "&& protocol=\"https\"" {
								urls = append(urls, "https://"+value[1])
							} else {
								urls = append(urls, "http://"+value[1])
							}
						}

					}
				} else {
					break
				}
			}
		}

	} else {
		for i := 1; i <= 20; i++ {
			url := fofa_api(keyword, fofa.Email, fofa.Fofa_token, i, 500)
			res := fofahttp(url, fofa.Fofa_timeout)
			if len(res.Results) > 0 {
				for _, value := range res.Results {
					if strings.Contains(value[1], "http") {
						urls = append(urls, value[1])
					} else {
						urls = append(urls, "http://"+value[1])
					}
				}
			}
		}
	}
	return
}

func Fafaall(keyword string) (urls []string) {
	color.RGBStyleFromString("244,211,49").Println("请耐心等待fofa搜索......")
	fofa := GetConfig()
	for i := 1; i <= 20; i++ {
		url := fofa_api(keyword, fofa.Email, fofa.Fofa_token, i, 500)
		res := fofahttp(url, fofa.Fofa_timeout)
		if len(res.Results) > 0 {
			for _, value := range res.Results {
				if strings.Contains(value[1], "http") {
					urls = append(urls, value[1])
				} else {
					urls = append(urls, "https://"+value[1])
				}
			}
		}
	}
	return
}

func Fofaall_out(keyword string) (result [][]string) {
	fofa := GetConfig()
	for i := 1; i <= 20; i++ {
		url := fofa_api(keyword, fofa.Email, fofa.Fofa_token, i, 500)
		res := fofahttp(url, fofa.Fofa_timeout)
		if len(res.Results) > 0 {
			result = append(result, res.Results...)
		} else {
			break
		}
	}
	return
}

func keyword_ips(filename string) (result []string) {
	var ips []string
	file, err := os.Open(filename)
	if err != nil {
		log.Println("Local file read error:", err)
		fmt.Println("[error] the input file is wrong!!!")
		os.Exit(1)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ips = append(ips, scanner.Text())
	}
	fmt.Println("共读取到" + strconv.Itoa(len(ips)) + "个ip")
	x := 100
	if len(ips) < x {
		s := ""
		for _, aa := range ips {
			s = s + " || " + "ip=\"" + aa + "\""
		}
		s = strings.Trim(strings.Trim(strings.Trim(s, " "), "|"), " ")
		result = append(result, s)
	} else {
		for i := 0; i+x < len(ips); i = i + x {
			s := ""
			for _, aa := range ips[i : i+x] {
				s = s + " || " + "ip=\"" + aa + "\""
			}
			s = strings.Trim(strings.Trim(strings.Trim(s, " "), "|"), " ")
			result = append(result, s)
		}
		i := len(ips) % x
		if i != 0 {
			s := ""
			for _, aa := range ips[len(ips)-i:] {
				s = s + " || " + "ip=\"" + aa + "\""
			}
			s = strings.Trim(strings.Trim(strings.Trim(s, " "), "|"), " ")
			result = append(result, s)
		}
	}
	return
}

func Fafaips_out(filename string) [][]string {
	fmt.Println("开始使用fofa批量搜索ip,请耐心等待....")
	keys := keyword_ips(filename)
	var results [][]string
	for _, x := range keys {
		result := Fofaall_out(x)
		results = append(results, result...)
	}
	fmt.Println("共收集" + strconv.Itoa(len(results)) + "条数据，已保存！！")
	return results
}
