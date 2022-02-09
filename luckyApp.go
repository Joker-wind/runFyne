package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
	"github.com/thedevsaddam/gojsonq"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

//设置中文字体
func init() {
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

// 获取历史记录
func getHistory(m *[][]string, n int) {
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

// 计数并排序
func parse(m *[][]string, cf map[string]int, cb map[string]int, f *[]Count, b *[]Count) {
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

// 通过map生成排序后的列表
func assort(c map[string]int, f *[]Count) {

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

// 完全随机
func getRandom(f []Count, b []Count, index int) (result string) {
	newF := make([]Count, len(f))
	newB := make([]Count, len(b))
	copy(newF, f)
	copy(newB, b)
	result = ""
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < index; i++ {
		fStr := make([]string, 0)
		bStr := make([]string, 0)
		// 生成前区5个号码
		for len(fStr) < 5 {
			a := rand.Intn(len(newF))
			fStr = append(fStr, newF[a].Key)
			newF = append(newF[:a], newF[a+1:]...)
		}
		// 生成后区2个号码
		for len(bStr) < 2 {
			if len(newB) == 0 {
				newB = make([]Count, len(b))
				copy(newB, b)
			}
			c := rand.Intn(len(newB))
			bStr = append(bStr, newB[c].Key)
			newB = append(newB[:c], newB[c+1:]...)
		}
		// 字符串升序
		sort.Slice(fStr, func(i, j int) bool {
			return strings.Compare(fStr[i], fStr[j]) < 0
		})
		sort.Slice(bStr, func(i, j int) bool {
			return strings.Compare(bStr[i], bStr[j]) < 0
		})
		result += " " + strings.Join(fStr, " ") + " " + strings.Join(bStr, " ") + "\n"
	}
	return result
}

func setTheme(ty string, A fyne.App) {
	// 设置主题模式
	if ty == "Light" {
		A.Settings().SetTheme(theme.LightTheme())
	} else {
		A.Settings().SetTheme(theme.DarkTheme())
	}
}

// 展示并添加组件
func showFSelectGroup(app fyne.App) {
	myWin := app.NewWindow("透乐大")
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
	themeLight := fyne.NewMenuItem("Light", func() {
		setTheme("Light", app)
	})
	themeDark := fyne.NewMenuItem("Dark", func() {
		setTheme("Dark", app)
	})
	setting := fyne.NewMenuItem("设置", nil)
	menuMain := fyne.NewMenu("菜单", setting)
	menuSet := fyne.NewMenu("主题", themeLight, themeDark)
	mainMenu := fyne.NewMainMenu(menuMain, menuSet)
	myWin.SetMainMenu(mainMenu)

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

	// 全局随机
	random := container.NewVBox()
	randText := widget.NewTextGrid()
	ct6 := container.New(layout.NewCenterLayout())
	ct6.Add(randText)
	randEntry := widget.NewEntry()
	randEntry.SetPlaceHolder("请输入数字1~7")
	randBtn := widget.NewButton("生成号码", func() {
		text := randEntry.Text
		n, _ := strconv.ParseInt(text, 10, 0)
		if text == "" {
			n = 7
		}
		if n <= 0 || n > 7 {
			randEntry.SetPlaceHolder("请输入数字1~7")
		} else {
			randText.SetText(getRandom(frontList, backList, int(n)))
		}
	})
	random.Add(widget.NewLabel("输入需要的组数，随机所有号码，默认生成7组。"))
	random.Add(randEntry)
	random.Add(randBtn)
	random.Add(ct6)

	// 选中随机
	ft := make([]string, 0)
	bt := make([]string, 0)
	mapF := make(map[string]int, 35)
	mapB := make(map[string]int, 12)
	for _, v := range frontList {
		ft = append(ft, v.Key)
	}
	for _, v := range backList {
		bt = append(bt, v.Key)
	}
	selectBox := container.NewVBox()
	selectBox.Add(widget.NewLabel("前区号码"))
	ct3 := container.New(layout.NewGridLayout(7))
	for _, v := range ft {
		tmp := v
		ct3.Add(widget.NewCheck(tmp, func(b bool) {
			if b {
				mapF[tmp] = 1
			} else {
				delete(mapF, tmp)
			}
		}))
	}
	selectBox.Add(ct3)
	selectBox.Add(widget.NewLabel("后区号码"))
	ct4 := container.New(layout.NewGridLayout(7))
	for _, v := range bt {
		tmp := v
		ct4.Add(widget.NewCheck(tmp, func(b bool) {
			if b {
				mapB[tmp] = 1
			} else {
				delete(mapB, tmp)
			}
		}))
	}
	ct5 := container.New(layout.NewCenterLayout())
	selectText := widget.NewTextGrid()
	ct5.Add(selectText)
	selectBox.Add(ct4)
	selectBtn := widget.NewButton("生成号码", func() {
		rest := selectRandom(mapF, mapB, ft, bt)
		selectText.SetText(rest)
	})
	selectBox.Add(selectBtn)
	selectBox.Add(ct5)

	// 将容器加入tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("历史记录", historyList),
		container.NewTabItem("分布统计", content),
		container.NewTabItem("全局随机", random),
		container.NewTabItem("选中随机", selectBox),
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	myWin.SetContent(tabs)
	myWin.Resize(fyne.NewSize(800, 600))
	myWin.ShowAndRun()
}

// 选中随机
func selectRandom(f, b map[string]int, fa, ba []string) string {
	// 选中的号码
	fList := make([]string, 0)
	bList := make([]string, 0)
	for k := range f {
		fList = append(fList, k)
	}
	for k := range b {
		bList = append(bList, k)
	}
	rand.Seed(time.Now().UnixNano())
	fAll := make([]string, 35)
	bAll := make([]string, 12)
	bAll2 := make([]string, 12)
	// 所有的号码
	copy(fAll, fa)
	copy(bAll, ba)
	copy(bAll2, ba)
	results := ""
	for p := 0; p < 7; p++ {
		fn, bn := len(fList), len(bList)
		fStr := make([]string, 0)
		bStr := make([]string, 0)
		// fn刚好5个时直接就是前区号码
		if fn == 5 {
			fStr = fList
			copy(fStr, fList)
			fList = []string{}
		} else if fn < 5 {
			// 先加选中数组
			fStr = append(fStr, fList...)
			// 清空所有数组
			fList, fAll = rmList(fList, fAll)
			// 保存不足5个时随机的数组
			resF := make([]string, 5-fn)
			// 获取随机结果和剩余总数组
			resF, fAll = randomList(5-fn, fAll)
			// 保存剩余数组
			fStr = append(fStr, resF...)
		} else if fn > 5 {
			resF := make([]string, 0)
			bakF := make([]string, 0)
			copy(bakF, fList)
			bakF, fAll = rmList(bakF, fAll)
			resF, fList = randomList(5, fList)
			fStr = append(fStr, resF...)
		}
		// 最后一次循环时bAll是空列表了
		if len(bAll) == 0 {
			bAll = bAll2
		}
		// bn刚好2个时直接就是后区号码
		if bn == 2 {
			bStr = bList
			copy(bStr, bList)
			bList = []string{}
		} else if bn < 2 {
			bStr = append(bStr, bList...)
			bList, bAll = rmList(bList, bAll)
			resB := make([]string, 2-bn)
			resB, bAll = randomList(2-bn, bAll)
			bStr = append(bStr, resB...)
		} else if bn > 2 {
			resB := make([]string, 0)
			bakB := make([]string, 0)
			copy(bakB, bList)
			bakB, bAll = rmList(bakB, bAll)
			resB, bList = randomList(2, bList)
			bStr = append(bStr, resB...)
		}

		// 字符串升序
		sort.Slice(fStr, func(i, j int) bool {
			return strings.Compare(fStr[i], fStr[j]) < 0
		})
		sort.Slice(bStr, func(i, j int) bool {
			return strings.Compare(bStr[i], bStr[j]) < 0
		})
		results += " " + strings.Join(fStr, " ") + " " + strings.Join(bStr, " ") + "\n"
	}
	return results
}

// 传入列表n,m 返回m-n
func rmList(t, a []string) (t2, a2 []string) {
	for _, v := range t {
		for j, k := range a {
			if v == k {
				a = append(a[:j], a[j+1:]...)
			}
		}
	}
	t = []string{}
	return t, a
}

// 返回随机结果和源列表
func randomList(n int, l []string) (resultList, l2 []string) {
	// 需要随机的数量刚好等于列表剩余元素的数量，直接返回
	if n == len(l) {
		resultList = make([]string, n)
		copy(resultList, l)
		l = []string{}
		return resultList, l
	}
	// 随机取数并删除取出的元素
	for i := 0; i < n; i++ {
		a := rand.Intn(len(l))
		resultList = append(resultList, l[a])
		l = append(l[:a], l[a+1:]...)
	}
	return resultList, l
}

//打包命令：fyne package -os windows -icon lucky.png
func main() {
	myApp := app.New()
	showFSelectGroup(myApp)

	defer func() {
		if err := os.Unsetenv("FYNE_FONT"); err != nil {
			log.Println("取消字体全局变量异常")
		}
	}()
}
