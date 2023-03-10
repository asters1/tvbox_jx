package js

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

func Init(vm *otto.Otto) *otto.Otto {
	vm = otto.New()
	//获取md5
	vm.Set("go_md5", func(call otto.FunctionCall) otto.Value {
		str, _ := call.Argument(0).ToString()
		data := []byte(str) //切片
		has := md5.Sum(data)
		md5str := fmt.Sprintf("%x", has) //将[]byte转成16进制
		vm_md5, _ := vm.ToValue(md5str)
		return vm_md5

	})
	//获取时间戳
	vm.Set("go_getTime", func(call otto.FunctionCall) otto.Value {
		i, _ := call.Argument(0).ToInteger()
		if i > 19 {
			i = 19
		}
		timeUnixNano := time.Now().UnixNano()
		str_time := strconv.FormatInt(timeUnixNano, 10)
		s, _ := vm.ToValue(str_time[:i])
		return s
	})
	//发送请求
	vm.Set("go_RequestClient", func(call otto.FunctionCall) otto.Value {
		FormatStr := func(jsonstr string) map[string]string {
			DataMap := make(map[string]string)
			Nslice := strings.Split(jsonstr, "\n")
			for i := 0; i < len(Nslice); i++ {
				if strings.Index(Nslice[i], ":") != -1 {
					a := Nslice[i][:strings.Index(Nslice[i], ":")]
					b := Nslice[i][strings.Index(Nslice[i], ":")+1:]
					DataMap[a] = b
				}
			}
			return DataMap

		}

		URL, _ := call.Argument(0).ToString()
		METHOD, _ := call.Argument(1).ToString()
		HEADER, _ := call.Argument(2).ToString()
		DATA, _ := call.Argument(3).ToString()

		URL = strings.TrimSpace(URL)
		METHOD = strings.TrimSpace(METHOD)
		HEADER = strings.TrimSpace(HEADER)
		DATA = strings.TrimSpace(DATA)
		if URL == "" || METHOD == "" {
			return otto.Value{}
		}

		HeaderMap := FormatStr(HEADER)
		DataMap := FormatStr(DATA)
		client := &http.Client{}
		if METHOD == "get" {
			METHOD = http.MethodGet
		} else if METHOD == "post" {
			METHOD = http.MethodPost

		}
		FormatData := ""
		for i, j := range DataMap {
			FormatData = FormatData + i + "=" + j + "&"
		}
		if FormatData != "" {
			FormatData = FormatData[:len(FormatData)-1]
		}
		requset, _ := http.NewRequest(
			METHOD,
			URL,
			strings.NewReader(FormatData),
		)
		if METHOD == http.MethodPost && requset.Header.Get("Content-Type") == "" {
			requset.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
		requset.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.71 Safari/537.36")
		for i, j := range HeaderMap {
			requset.Header.Set(i, j)
		}
		resp, _ := client.Do(requset)
		body_bit, _ := ioutil.ReadAll(resp.Body)
		headerMap := resp.Header
		jsonByte, err := json.Marshal(headerMap)
		if err != nil {
			fmt.Printf("Marshal with error: %+v\n", err)
		}
		header := string(jsonByte)

		defer resp.Body.Close()
		status := strconv.Itoa(resp.StatusCode)
		body := string(body_bit)
		res_str := make(map[string]string)

		res_str["status"] = status
		res_str["header"] = header
		res_str["body"] = body
		result, _ := vm.ToValue(res_str)

		return result
	})

	return vm
}
