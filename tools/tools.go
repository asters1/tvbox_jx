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

	LogPrintln_shang(startTime, "正在加载BaseUrl...")
	sourceBaseUrl := gjson.Get(SourceJson, sid+".sourceUrl").String()
	LogPrintln_xia(startTime, sourceBaseUrl)
	vm.Set("sourceBaseUrl", sourceBaseUrl)
	LogPrintln_shang(startTime, "正在加载BaseHeader...")
	sourceBaseHeader := gjson.Get(SourceJson, sid+".sourceBaseHeader").String()
	LogPrintln_xia(startTime, sourceBaseHeader)
	vm.Set("sourceBaseHeader", sourceBaseHeader)
	LogPrintln_shang(startTime, "正在加载searchUrl...")
	sourceSUrl := gjson.Get(SourceJson, sid+".searchUrl").String()
	sourceSearchUrl := ReplaceKey(sourceSUrl, keyword)
	sourceSearchUrl = CheckUrl(sourceBaseUrl, sourceSearchUrl)
	LogPrintln_xia(startTime, sourceSearchUrl)
	vm.Set("sourceSearchUrl", sourceSearchUrl)
	LogPrintln_shang(startTime, "正在加载searchMethod...")
	sourceSearchMethod := gjson.Get(SourceJson, sid+".searchMethod").String()
	LogPrintln_xia(startTime, sourceSearchMethod)
	vm.Set("sourceSearchMethod", sourceSearchMethod)

	LogPrintln_shang(startTime, "正在加载searchHeader...")
	sourceSearchHeader := gjson.Get(SourceJson, sid+".searchHeader").String()
	sourceSearchHeader = ReplaceKey(sourceSearchHeader, keyword)
	LogPrintln_xia(startTime, sourceSearchHeader)
	vm.Set("sourceSearchHeader", sourceBaseHeader+"\n"+sourceSearchHeader)
	LogPrintln_shang(startTime, "正在加载searchData...")
	sourceSearchData := gjson.Get(SourceJson, sid+".searchData").String()
	sourceSearchData = ReplaceKey(sourceSearchData, keyword)
	LogPrintln_xia(startTime, sourceSearchData)
	vm.Set("sourceSearchData", sourceSearchData)

	LogPrintln_sanjao(startTime, "开始搜索关键字:"+key)
	vm.Run(`
	searchResult=go_RequestClient(sourceSearchUrl,sourceSearchMethod,sourceSearchHeader,sourceSearchData)
	resultBody=searchResult.body
	`)
	res, _ := vm.Get("resultBody")
	result := res.String()
	LogPrintln_jtx(startTime, "开始解析搜索页")
	ress := JxResult_slice(vm, result, "")
	fmt.Println(ress)

}

func JxResult_string(vm *otto.Otto, jstr string, rule string) string {
	rule = strings.TrimSpace(rule)

	if strings.HasPrefix(rule, "@json:") {
		rule = rule[6:]
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		res := gjson.Get(jstr, rule).String()
		return res
	} else if strings.HasPrefix(rule, "@xpath:") {
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		rule = rule[7:]
		doc, _ := htmlquery.Parse(strings.NewReader(jstr))
		nodes, _ := htmlquery.Query(doc, rule)
		result := htmlquery.InnerText(nodes)

		return result
	} else if strings.HasPrefix(rule, "@js:") {
		rule = rule[4:]
		a, _ := vm.Run(rule)
		result := a.String()
		return result
	} else if strings.HasPrefix(rule, "@re:") {
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		rule = rule[4:]
		rule = strings.TrimSpace(rule)
		re := regexp.MustCompile(rule)
		res := re.FindStringSubmatch(jstr)
		if len(res) > 0 {
			return res[1]

		}
		return ""

	}

	return ""
}
func JxResult_slice(vm *otto.Otto, jstr string, rule string) []string {
	rule = ` @js: 

	var a=["a","b","c"]
	return a
	`
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
		fmt.Println(htmlquery.InnerText(nodes[0]))
		for i := 0; i < len(nodes); i++ {
			result = append(result, htmlquery.InnerText(nodes[i]))

		}

		return result
	} else if strings.HasPrefix(rule, "@js:") {
		rule = rule[4:]
		a, _ := vm.Run(rule)
		result := a
		fmt.Println(result)
		return nil
	} else if strings.HasPrefix(rule, "@re:") {
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		rule = rule[4:]
		rule = strings.TrimSpace(rule)
		re := regexp.MustCompile(rule)
		res := re.FindStringSubmatch(jstr)
		if len(res) > 0 {
			return nil

		}
		return nil

	}

	return nil
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
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)
	fmt.Println(" " + str)
}
func LogPrintln_sanjao(old_time int64, str string) {
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)
	fmt.Println(" ➤➤➤ " + str)
}
func LogPrintln_shang(old_time int64, str string) {
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)
	fmt.Println("「   " + str)
}
func LogPrintln_xia(old_time int64, str string) {
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)
	fmt.Println(" └   " + str)
}
func LogPrintln_jtx(old_time int64, str string) {
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)
	fmt.Println(" ⬇   " + str)
}
func LogPrintln_jts(old_time int64, str string) {
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)
	fmt.Println(" ⬆   " + str)
}
