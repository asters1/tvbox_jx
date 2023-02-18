package tools

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/robertkrimen/otto"
	"github.com/tidwall/gjson"
)

func Spider(vm *otto.Otto, sid string, key string, spath string) {
	keyword := url.QueryEscape(key)
	startTime := time.Now().UnixNano() / 1e6

	SourceJson, err := ReadSourceFile(spath)

	if err != nil {
		fmt.Println("读取[" + spath + "]文件失败!请检查!!!")

	}
	sourceName := gjson.Get(SourceJson, sid+".sourceName").String()
	LogPrintln_sanjao(startTime, "开始测试源:"+sourceName)

	sourceBaseUrl := gjson.Get(SourceJson, sid+".sourceUrl").String()
	vm.Set("sourceBaseUrl", sourceBaseUrl)
	sourceBaseHeader := gjson.Get(SourceJson, sid+".sourceBaseHeader").String()
	vm.Set("sourceBaseHeader", sourceBaseHeader)
	sourceSUrl := gjson.Get(SourceJson, sid+".searchUrl").String()
	sourceSearchUrl := ReplaceKey(sourceSUrl, keyword)
	sourceSearchUrl = CheckUrl(sourceBaseUrl, sourceSearchUrl)
	vm.Set("sourceSearchUrl", sourceSearchUrl)
	sourceSearchMethod := gjson.Get(SourceJson, sid+".searchMethod").String()
	vm.Set("sourceSearchMethod", sourceSearchMethod)

	sourceSearchHeader := gjson.Get(SourceJson, sid+".searchHeader").String()
	sourceSearchHeader = ReplaceKey(sourceSearchHeader, keyword)
	vm.Set("sourceSearchHeader", sourceBaseHeader+"\n"+sourceSearchHeader)
	sourceSearchData := gjson.Get(SourceJson, sid+".searchData").String()
	sourceSearchData = ReplaceKey(sourceSearchData, keyword)
	vm.Set("sourceSearchData", sourceSearchData)

	LogPrintln_sanjao(startTime, "开始搜索关键字:"+key)
	vm.Run(`
	searchResult=go_RequestClient(sourceSearchUrl,sourceSearchMethod,sourceSearchHeader,sourceSearchData)
	resultBody=searchResult.body
	`)
	res_body, err := vm.Get("resultBody")
	if err != nil {
		LogPrintln_err(startTime, "获取失败!!!"+sourceSearchUrl)
		return
	}
	LogPrintln_success(startTime, "获取成功:"+sourceSearchUrl)
	result := res_body.String()
	LogPrintln_jtx(startTime, "开始解析搜索页")
	searchVideoList := gjson.Get(SourceJson, sid+".searchVideoList").String()
	searchVideoListResult := JxResult_slice(vm, result, searchVideoList)

	LogPrintln_shang(startTime, "获取视频列表")
	if len(searchVideoListResult) > 0 {
		LogPrintln_xia(startTime, "列表大小:"+strconv.Itoa(len(searchVideoListResult)))
	} else {

		LogPrintln_xia(startTime, "视频列表为空")
	}
	videoInfo := SelectVideo(0, searchVideoListResult)
	sourceSearchVideoName := gjson.Get(SourceJson, sid+".searchVideoName").String()
	LogPrintln_shang(startTime, "获取视频名")
	videoName := JxResult_string(vm, videoInfo, sourceSearchVideoName)
	LogPrintln_xia(startTime, videoName)
	sourceSearchVideoAuthor := gjson.Get(SourceJson, sid+".searchVideoAuthor").String()
	LogPrintln_shang(startTime, "导演")
	SearchVideoAuthor := JxResult_string(vm, videoInfo, sourceSearchVideoAuthor)
	LogPrintln_xia(startTime, SearchVideoAuthor)

}

func SelectVideo(index int, list []string) string {
	return list[index]

}

func JxResult_string(vm *otto.Otto, jstr string, rule string) string {
	rule = strings.TrimSpace(rule)

	if strings.HasPrefix(rule, "@json:") {
		rule = rule[6:]
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		res := gjson.Get(jstr, rule).String()
		return strings.TrimSpace(res)
	} else if strings.HasPrefix(rule, "@xpath:") {
		rule = strings.ReplaceAll(rule, "\n", "")

		rule = strings.TrimSpace(rule)
		rule = rule[7:]
		doc, _ := htmlquery.Parse(strings.NewReader(jstr))
		nodes, _ := htmlquery.Query(doc, rule)
		result := htmlquery.InnerText(nodes)

		return strings.TrimSpace(result)
	} else if strings.HasPrefix(rule, "@js:") {
		rule = rule[4:]
		vm.Run(rule)
		a, _ := vm.Get("result")
		result := a.String()
		return strings.TrimSpace(result)
	} else if strings.HasPrefix(rule, "@re:") {
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		rule = rule[4:]
		rule = strings.TrimSpace(rule)
		re := regexp.MustCompile(rule)
		res := re.FindStringSubmatch(jstr)
		if len(res) > 1 {
			return strings.TrimSpace(res[1])

		}
		return ""

	}

	return "格式有误，请检查!"
}
func JxResult_slice(vm *otto.Otto, jstr string, rule string) []string {
	rule = strings.TrimSpace(rule)
	if strings.HasPrefix(rule, "@json:") {
		rule = rule[6:]
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		res := gjson.Get(jstr, rule).Array()
		var result []string
		for i := 0; i < len(res); i++ {
			result = append(result, res[i].String())
		}

		return result
	} else if strings.HasPrefix(rule, "@xpath:") {
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		rule = rule[7:]
		doc, _ := htmlquery.Parse(strings.NewReader(jstr))
		nodes, _ := htmlquery.QueryAll(doc, rule)
		//nodes, _ := htmlquery.QueryAll(doc, rule)
		//fmt.Println(jstr)
		var result []string
		for i := 0; i < len(nodes); i++ {
			result = append(result, htmlquery.InnerText(nodes[i]))
		}
		return result

	} else if strings.HasPrefix(rule, "@js:") {
		rule = rule[4:]
		vm.Run(rule)
		a, _ := vm.Get("result")
		var result []string
		for i := 0; i < len(a.Object().Keys()); i++ {
			res, _ := a.Object().Get(strconv.Itoa(i))
			result = append(result, res.String())

		}
		return result
	} else if strings.HasPrefix(rule, "@re:") {

		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		rule = rule[4:]
		rule = strings.TrimSpace(rule)
		re := regexp.MustCompile(rule)
		res := re.FindAllStringSubmatch(jstr, -1)

		var result []string
		for i := 0; i < len(res); i++ {
			if len(res[i]) > 1 {

				result = append(result, res[i][1])
			}
		}

		return result
	}
	var result []string
	result = append(result, "格式有误，请检查!")

	return result
}

func ReadSourceFile(path string) (string, error) {

	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("读取[" + path + "]文件失败!请检查!!!")

		return "", err
	}
	sourceJsonStr := string(content)
	return sourceJsonStr, nil
}
func ReplaceKey(str string, key string) string {
	strs := strings.ReplaceAll(str, "{{key}}", key)
	return strs
}
func CheckUrl(baseUrl string, url string) string {
	if !strings.HasPrefix(url, "http") {
		return baseUrl + url
	}
	return url

}

func LogPrintln(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" " + str)
}
func LogPrintln_sanjao(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" ➤➤➤ " + str)
}
func LogPrintln_shang(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println("「   " + str)
}
func LogPrintln_xia(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" └   " + str)
}
func LogPrintln_jtx(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" ⬇   " + str)
}
func LogPrintln_jts(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" ⬆   " + str)
}
func LogPrintln_err(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" X   " + str)
}
func LogPrintln_success(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" ✔   " + str)
}
func LogTime(old_time int64) {
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)

}
