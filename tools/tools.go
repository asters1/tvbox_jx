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
	startTime := time.Now().UnixNano() / 1e6

	SourceJson, err := ReadSourceFile(spath)

	if err != nil {
		fmt.Println("读取[" + spath + "]文件失败!请检查!!!")

	}
	//JS函数初始化
	jsinit := gjson.Get(SourceJson, sid+".jsInit").String()
	JxResult_string(vm, "", jsinit)

	//源名称
	sourceName := gjson.Get(SourceJson, sid+".sourceName").String()
	LogPrintln_sanjao(startTime, "开始测试源:"+sourceName)
	//基础URL
	sourceBaseUrl := gjson.Get(SourceJson, sid+".sourceUrl").String()
	vm.Set("sourceBaseUrl", sourceBaseUrl)
	//基础Header
	sourceBaseHeader := gjson.Get(SourceJson, sid+".sourceBaseHeader").String()
	sourceBaseHeader = JxResult_string(vm, "", sourceBaseHeader)
	vm.Set("sourceBaseHeader", sourceBaseHeader)
	//搜索
	DetailUrl := SearchSpider(startTime, SourceJson, sid, key, sourceBaseUrl, sourceBaseHeader, vm)
	//详情页
	DetailSpider(startTime, SourceJson, sid, key, DetailUrl, sourceBaseHeader, vm)

}
func GetReturnString(startTime int64, vm *otto.Otto, pstr string, sid string, source_jstr string, key string, jx_string string) string {
	LogPrintln_shang(startTime, pstr)
	value := gjson.Get(source_jstr, sid+"."+key).String()
	result := JxResult_string(vm, jx_string, value)
	LogPrintln_xia(startTime, result)
	return result

}

func SelectVideo(index int, list []string) string {
	return list[index]

}

func JxResult_string(vm *otto.Otto, jstr string, rule string) string {
	rule = strings.TrimSpace(rule)

	rule_js := ""
	if strings.HasPrefix(rule, "@json:") {
		rule = rule[6:]
		if strings.Contains(rule, "@js:") {
			rule_js = rule[strings.Index(rule, "@js:"):]
			rule = rule[:strings.Index(rule, "@js:")]
		}
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		res := gjson.Get(jstr, rule).String()

		if rule_js != "" {
			vm.Set("result", strings.TrimSpace(res))
			rule_js = rule_js[4:]
			vm.Run(rule_js)
			a, _ := vm.Get("result")
			result := a.String()
			return strings.TrimSpace(result)
		} else {
			return strings.TrimSpace(res)
		}
	} else if strings.HasPrefix(rule, "@xpath:") {
		rule = strings.ReplaceAll(rule, "\n", "")

		rule = strings.TrimSpace(rule)
		rule = rule[7:]
		if strings.Contains(rule, "@js:") {
			rule_js = rule[strings.Index(rule, "@js:"):]
			rule = rule[:strings.Index(rule, "@js:")]
		}
		doc, _ := htmlquery.Parse(strings.NewReader(jstr))
		nodes, _ := htmlquery.Query(doc, rule)
		result := htmlquery.InnerText(nodes)

		if rule_js != "" {
			vm.Set("result", strings.TrimSpace(result))
			rule_js = rule_js[4:]
			vm.Run(rule_js)
			a, _ := vm.Get("result")
			result := a.String()
			return strings.TrimSpace(result)
		} else {
			return strings.TrimSpace(result)
		}
	} else if strings.HasPrefix(rule, "@js:") {
		rule = rule[4:]
		vm.Run(rule)
		a, _ := vm.Get("result")
		result := a.String()
		return strings.TrimSpace(result)
	} else if strings.HasPrefix(rule, "@re:") {
		rule = strings.TrimSpace(rule)
		rule = rule[4:]
		if strings.Contains(rule, "@js:") {
			rule_js = rule[strings.Index(rule, "@js:"):]
			rule = rule[:strings.Index(rule, "@js:")]
		}
		rule = strings.TrimSpace(rule)
		re := regexp.MustCompile(rule)
		res := re.FindStringSubmatch(jstr)
		if len(res) > 1 {
			if rule_js != "" {
				vm.Set("result", strings.TrimSpace(res[1]))
				rule_js = rule_js[4:]
				vm.Run(rule_js)
				a, _ := vm.Get("result")
				result := a.String()
				return strings.TrimSpace(result)
			} else {
				return strings.TrimSpace(res[1])
			}

		}
		return ""

	} else if rule != "" {

		return "格式有误，请检查!"
	} else if strings.HasPrefix(rule, "@str:") {
		rule = rule[5:]
		rule = strings.TrimSpace(rule)
		return rule

	}
	return ""

}

func JxResult_slice(vm *otto.Otto, jstr string, rule string) []string {
	rule = strings.TrimSpace(rule)
	rule_js := ""
	if strings.HasPrefix(rule, "@json:") {
		rule = rule[6:]
		if strings.Contains(rule, "@js:") {
			rule_js = rule[strings.Index(rule, "@js:"):]
			rule = rule[:strings.Index(rule, "@js:")]
		}
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		res := gjson.Get(jstr, rule).Array()
		var result []string
		for i := 0; i < len(res); i++ {
			result = append(result, res[i].String())
		}

		if rule_js != "" {
			vm.Set("result", result)
			rule_js = rule_js[4:]
			vm.Run(rule_js)
			a, _ := vm.Get("result")
			for i := 0; i < len(a.Object().Keys()); i++ {
				res, _ := a.Object().Get(strconv.Itoa(i))
				result = append(result, res.String())

			}
			return result
		}
		return result
	} else if strings.HasPrefix(rule, "@xpath:") {
		rule = strings.ReplaceAll(rule, "\n", "")
		rule = strings.TrimSpace(rule)
		rule = rule[7:]
		if strings.Contains(rule, "@js:") {
			rule_js = rule[strings.Index(rule, "@js:"):]
			rule = rule[:strings.Index(rule, "@js:")]
		}
		doc, _ := htmlquery.Parse(strings.NewReader(jstr))
		nodes, _ := htmlquery.QueryAll(doc, rule)
		//nodes, _ := htmlquery.QueryAll(doc, rule)
		//fmt.Println(jstr)
		var result []string
		for i := 0; i < len(nodes); i++ {
			result = append(result, htmlquery.InnerText(nodes[i]))
		}
		if rule_js != "" {
			vm.Set("result", result)
			rule_js = rule_js[4:]
			vm.Run(rule_js)
			a, _ := vm.Get("result")
			for i := 0; i < len(a.Object().Keys()); i++ {
				res, _ := a.Object().Get(strconv.Itoa(i))
				result = append(result, res.String())

			}
			return result
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

		rule = strings.TrimSpace(rule)
		rule = rule[4:]
		if strings.Contains(rule, "@js:") {
			rule_js = rule[strings.Index(rule, "@js:"):]
			rule = rule[:strings.Index(rule, "@js:")]
		}
		rule = strings.TrimSpace(rule)
		re := regexp.MustCompile(rule)
		res := re.FindAllStringSubmatch(jstr, -1)

		var result []string
		for i := 0; i < len(res); i++ {
			if len(res[i]) > 1 {

				result = append(result, res[i][1])
			}
		}
		if rule_js != "" {
			vm.Set("result", result)
			rule_js = rule_js[4:]
			vm.Run(rule_js)
			a, _ := vm.Get("result")
			for i := 0; i < len(a.Object().Keys()); i++ {
				res, _ := a.Object().Get(strconv.Itoa(i))
				result = append(result, res.String())

			}
			return result
		}
		return result

	} else if rule != "" {
		var result []string
		result = append(result, "格式有误，请检查!")

		return result
	} else if strings.HasPrefix(rule, "@strs:") {

		rule = rule[6:]
		rule = strings.TrimSpace(rule)
		if strings.HasPrefix(rule, "[") && strings.HasSuffix(rule, "]") {
			rule = rule[1 : len(rule)-1]
			result := strings.Split(rule, ",")
			return result

		} else {
			var result []string
			result = append(result, "格式有误，请检查!")

			return result

		}
	}
	var result []string
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
	fmt.Println(" ➤➤ " + str)
}
func LogPrintln_shang(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println("「  " + str)
}
func LogPrintln_xia(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" └  " + str)
}
func LogPrintln_jtx(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" ⬇  " + str)
}
func LogPrintln_jts(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" ⬆  " + str)
	fmt.Println()
}
func LogPrintln_err(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" X  " + str)
}
func LogPrintln_success(old_time int64, str string) {
	LogTime(old_time)
	fmt.Println(" ✔  " + str)
}
func LogTime(old_time int64) {
	now_time := time.Now().UnixNano() / 1e6
	a := now_time - old_time
	b, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(a)/float64(1000)), 64)
	fmt.Printf("[%.2fs]", b)

}
func SearchSpider(startTime int64, SourceJson string, sid string, key string, sourceBaseUrl string, sourceBaseHeader string, vm *otto.Otto) string {

	keyword := url.QueryEscape(key)
	vm.Set("keyword", keyword)
	vm.Set("key", key)
	//搜索URL
	sourceSUrl := gjson.Get(SourceJson, sid+".searchUrl").String()
	sourceSearchUrl := ReplaceKey(sourceSUrl, keyword)
	sourceSearchUrl = JxResult_string(vm, "", sourceSearchUrl)
	sourceSearchUrl = CheckUrl(sourceBaseUrl, sourceSearchUrl)
	vm.Set("sourceSearchUrl", sourceSearchUrl)
	//搜索方法
	sourceSearchMethod := gjson.Get(SourceJson, sid+".searchMethod").String()
	vm.Set("sourceSearchMethod", sourceSearchMethod)
	//搜索Header
	sourceSearchHeader := gjson.Get(SourceJson, sid+".searchHeader").String()
	sourceSearchHeader = JxResult_string(vm, "", sourceSearchHeader)
	sourceSearchHeader = ReplaceKey(sourceSearchHeader, keyword)
	vm.Set("sourceSearchHeader", sourceBaseHeader+"\n"+sourceSearchHeader)
	//搜索数据，post才会用到
	sourceSearchData := gjson.Get(SourceJson, sid+".searchData").String()
	sourceSearchData = ReplaceKey(sourceSearchData, keyword)
	vm.Set("sourceSearchData", sourceSearchData)

	LogPrintln_sanjao(startTime, "开始搜索关键字:"+key)
	vm.Run(`
	searchResult=go_RequestClient(sourceSearchUrl,sourceSearchMethod,sourceSearchHeader,sourceSearchData)
	searchBody=searchResult.body
	`)
	res_body, err := vm.Get("searchBody")
	if err != nil {
		LogPrintln_err(startTime, "获取失败!!!"+sourceSearchUrl)
		return ""
	}
	LogPrintln_success(startTime, "获取成功:"+sourceSearchUrl)
	result := res_body.String()
	//解析搜索页
	LogPrintln_jtx(startTime, "开始解析搜索页")
	searchVideoList := gjson.Get(SourceJson, sid+".searchVideoList").String()
	searchVideoListResult := JxResult_slice(vm, result, searchVideoList)

	//解析视频列表
	LogPrintln_shang(startTime, "获取视频列表")
	if len(searchVideoListResult) > 0 {
		LogPrintln_xia(startTime, "列表大小:"+strconv.Itoa(len(searchVideoListResult)))
	} else {

		LogPrintln_xia(startTime, "视频列表为空")
		LogPrintln_jts(startTime, "搜索解析完成")
		return ""
	}
	videoInfo := SelectVideo(0, searchVideoListResult)
	vm.Set("result", videoInfo)
	//视频信息列表
	GetReturnString(startTime, vm, "视频名称:", sid, SourceJson, "searchVideoName", videoInfo)
	//地区
	GetReturnString(startTime, vm, "地区:", sid, SourceJson, "searchVideoArea", videoInfo)
	//导演
	GetReturnString(startTime, vm, "导演:", sid, SourceJson, "searchVideoAuthor", videoInfo)
	//年份
	GetReturnString(startTime, vm, "年份:", sid, SourceJson, "searchVideoYear", videoInfo)
	//主演
	GetReturnString(startTime, vm, "主演:", sid, SourceJson, "searchVideoStarring", videoInfo)
	//类型
	GetReturnString(startTime, vm, "类型:", sid, SourceJson, "searchVideoKind", videoInfo)
	//最新章节
	GetReturnString(startTime, vm, "最新章节:", sid, SourceJson, "searchVideoLastChapter", videoInfo)
	//封面
	GetReturnString(startTime, vm, "封面:", sid, SourceJson, "searchVideoPic", videoInfo)
	//简介
	GetReturnString(startTime, vm, "简介:", sid, SourceJson, "searchVideoInfo", videoInfo)
	//详情页URL
	videoUrl := GetReturnString(startTime, vm, "详情页URL:", sid, SourceJson, "searchVideoUrl", videoInfo)
	LogPrintln_jts(startTime, "搜索解析完成")
	return videoUrl

}

func DetailSpider(startTime int64, SourceJson string, sid string, key string, detailUrl string, sourceDetailHeader string, vm *otto.Otto) string {
	LogPrintln_jtx(startTime, "开始请求详情页")
	//详情页URL
	vm.Set("sourceDetailUrl", detailUrl)
	//详情方法
	sourceDetaiMethod := gjson.Get(SourceJson, sid+".DetailMethod").String()
	vm.Set("sourceDetailMethod", sourceDetaiMethod)
	vm.Run(`
	console.log(sourceDetailMethod)`)

	//详情Header
	sourceDetaiHeader := gjson.Get(SourceJson, sid+".DetailHeader").String()
	sourceDetaiHeader = JxResult_string(vm, "", sourceDetaiHeader)
	vm.Set("sourceDetailHeader", sourceDetaiHeader+"\n"+sourceDetaiHeader)
	//搜索数据，post才会用到
	sourceDetaiData := gjson.Get(SourceJson, sid+".DetailData").String()
	vm.Set("sourceDetailData", sourceDetaiData)
	//"searchHeader": "@js:t=go_getTime(10) ;result=getHeaders(t,\"/api.php/provide/searchVideorealme4ac3fe96a6133de96904b8d3c8cfe16dRMX1931com.sevenVideo.app.android0101100021\"+keyword+t+\"XSpeUFjJ\")",

	vm.Run(`
	DetailResult=go_RequestClient(sourceDetailUrl,sourceDetailMethod,sourceDetailHeader,sourceDetailData)
	resultBody=DetailResult.body
	`)
	res_body, err := vm.Get("resultBody")
	if err != nil {
		LogPrintln_err(startTime, "获取失败!!!"+detailUrl)
		return ""
	}
	LogPrintln_success(startTime, "获取成功:"+detailUrl)
	result := res_body.String()
	fmt.Println(result)
	return ""
}
