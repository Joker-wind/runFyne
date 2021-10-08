package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"github.com/thedevsaddam/gojsonq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

func init() {
	//设置中文字体
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		//楷体:simkai.ttf
		//黑体:simhei.ttf
		if strings.Contains(path, "simkai.ttf") {
			err := os.Setenv("FYNE_FONT", path)
			if err != nil {
				log.Println("设置字体全局变量异常")
			}
			break
		}
	}
}

func getHistory(m *[][]string, n int) {
	// 获取历史记录
	client := &http.Client{
		Timeout: time.Second * 3,
	}
	url := "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListV1.qry?gameNo=85&provinceId=0&pageSize=" + fmt.Sprintf("%d", n) + "&isVerify=1&pageNo=1"
	//url := "https://webapi.sporttery.cn/gateway/lottery/getHistoryPageListV1.qry?gameNo=85&provinceId=0&pageSize=2&isVerify=1&pageNo=1"
	res, _ := http.NewRequest(http.MethodGet, url, nil)

	res.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	res.Header.Add("Accept-Encoding", "gzip, deflate, br")
	res.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	res.Header.Add("Connection", "keep-alive")
	res.Header.Add("Host", "webapi.sporttery.cn")
	res.Header.Add("Origin", "https://static.sporttery.cn")
	res.Header.Add("Referer", "https://static.sporttery.cn/")
	res.Header.Add("sec-ch-ua", "\"Not A;Brand\";v=\"99\", \"Chromium\";v=\"90\", \"Google Chrome\";v=\"90\"")
	res.Header.Add("sec-ch-ua-mobile", "?0")
	res.Header.Add("Sec-Fetch-Dest", "empty")
	res.Header.Add("Sec-Fetch-Mode", "cors")
	res.Header.Add("Sec-Fetch-Site", "same-site")
	res.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.72 Safari/537.36")

	resp, _ := client.Do(res)
	body, _ := ioutil.ReadAll(resp.Body)

	//defer resp.Body.Close()

	gq := gojsonq.New().FromString(string(body))
	list := gq.From("value.list").Select("lotteryDrawNum", "lotteryDrawResult").Get()
	gq.Reset()

	for _, e := range list.([]interface{}) {
		//e的类型和值：map[string]interface {},map[lotteryDrawNum:21112 lotteryDrawResult:18 21 22 23 35 11 12]
		Num := e.(map[string]interface{})["lotteryDrawNum"].(string)
		Result := e.(map[string]interface{})["lotteryDrawResult"].(string)
		cu := []string{Num, Result}
		//fmt.Println(Num, Result)
		*m = append(*m, cu)
	}
}

func parse(m *[][]string, cf map[string]int, cb map[string]int, f *[]Count, b *[]Count) {
	// 计数并排序
	for _, values := range *m {
		num := values[1]
		arr := strings.Split(num, " ")
		for _, v := range arr[:5] {
			if _, ok := cf[string(v)]; ok {
				cf[string(v)] += 1
			} else {
				cf[string(v)] = 1
			}
		}
		for _, v := range arr[5:] {
			if _, ok := cb[string(v)]; ok {
				cb[string(v)] += 1
			} else {
				cb[string(v)] = 1
			}
		}
	}
	assort(cf, f)
	assort(cb, b)

}

func assort(c map[string]int, f *[]Count) {
	// 通过map生成排序后的列表

	for k, v := range c {
		*f = append(*f, Count{k, v})
	}
	fs := CsSort{
		CountList: *f,
		less: func(x, y Count) bool {
			// 按照key升序
			return strings.Compare(x.Key, y.Key) < 0
			// 按照value升序
			//return x.Value < y.Value
		},
	}
	sort.Sort(fs)
	*f = fs.CountList
}

type Count struct {
	Key   string
	Value int
}

type CsSort struct {
	CountList []Count
	less      func(x, y Count) bool
}

func (c CsSort) Len() int {
	return len(c.CountList)
}

func (c CsSort) Less(x, y int) bool {
	return c.less(c.CountList[x], c.CountList[y])
}

func (c CsSort) Swap(i, j int) {
	c.CountList[i], c.CountList[j] = c.CountList[j], c.CountList[i]
}

func main() {
	myApp := app.New()
	myWin := myApp.NewWindow("透乐大")

	sMap := make([][]string, 0)
	fMap := make(map[string]int)
	bMap := make(map[string]int)
	frontList := make([]Count, 0)
	backList := make([]Count, 0)
	// 查询记录数
	n := 100
	getHistory(&sMap, n)
	parse(&sMap, fMap, bMap, &frontList, &backList)

	// 菜单
	menuItem := fyne.NewMenuItem("别点", func() {
		nt := fyne.NewNotification("警告", "你已经被包围了！请放下服务器！")
		myApp.SendNotification(nt)
	})
	menu := fyne.NewMenu("菜单", menuItem)
	mainMenu := fyne.NewMainMenu(menu)
	myWin.SetMainMenu(mainMenu)

	//log.Println(frontList)
	//log.Println(backList)
	// 展示历史记录
	historyList := widget.NewTable(
		func() (int, int) {
			return len(sMap), 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(sMap[i.Row][i.Col])
		})
	historyList.SetColumnWidth(1, 170)

	// 展示分布统计
	content := container.NewVBox()
	content.Add(widget.NewLabel(fmt.Sprintf("以下数据统计号码在最近%d期的占比（0~%d）", n, n)))

	content.Add(widget.NewLabel("前区号码"))
	ct1 := container.New(layout.NewGridLayout(6))
	for _, v := range frontList {
		ct := container.NewHBox()
		bar := widget.NewProgressBar()
		bar.Min = 0
		bar.Max = float64(n)
		bar.Value = float64(v.Value)
		ct.Add(widget.NewLabel(v.Key))
		ct.Add(bar)
		ct1.Add(ct)
	}
	content.Add(ct1)

	content.Add(widget.NewLabel("后区号码"))
	ct2 := container.New(layout.NewGridLayout(6))
	for _, v := range backList {
		ct := container.NewHBox()
		bar := widget.NewProgressBar()
		bar.Min = 0
		bar.Max = float64(n)
		bar.Value = float64(v.Value)
		ct.Add(widget.NewLabel(v.Key))
		ct.Add(bar)
		ct2.Add(ct)
	}
	content.Add(ct2)

	// 将容器加入tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("历史记录", historyList),
		container.NewTabItem("分布统计", content),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	myWin.SetContent(tabs)
	myWin.Resize(fyne.NewSize(800, 600))
	myWin.ShowAndRun()

	if err := os.Unsetenv("FYNE_FONT"); err != nil {
		log.Println("取消字体全局变量异常")
	}
}
