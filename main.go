package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var usageTemplate = `usage:
	<py> -l
		List index of classes
	<py> -i [<start> <end>]
		interactive mode
	<py> -d <index> 
		download mails using default date setting
		FROM %s TO %s
	<py> -s <start> <end> <index>
		download mails using date setting by hand. 
		eg: <py> -s %s %s 0
	<py> -h
		show this
`

func setConfigs() bool {
	classOptions := ReadConfigDir("./configsdir")

	dateBegin, dateEnd := GetDateRange()
	nIndexChoice := -1

	if len(os.Args) == 3 && os.Args[1] == "-d" {
		MailFetchConfig.DateStart, MailFetchConfig.DateEnd =
			dateBegin, dateEnd
		nIndexChoice, _ = strconv.Atoi(os.Args[2])
		//检查范围
		if nIndexChoice < 0 || nIndexChoice > len(classOptions)-1 {
			fmt.Println("Index out of range.")
			return false
		}
		//选择设置
		MailFetchConfig = classOptions[nIndexChoice]
	} else if len(os.Args) == 5 && os.Args[1] == "-s" {
		dateBegin, _ = time.ParseInLocation("200601021504", os.Args[2], time.Local)
		dateEnd, _ = time.ParseInLocation("200601021504", os.Args[3], time.Local)

		nIndexChoice, _ = strconv.Atoi(os.Args[4])
		//检查范围
		if nIndexChoice < 0 || nIndexChoice > len(classOptions)-1 {
			fmt.Println("Index out of range.")
			return false
		}

		//选择设置
		MailFetchConfig = classOptions[nIndexChoice]
	} else if len(os.Args) == 2 && os.Args[1] == "-l" {
		for i, item := range classOptions {
			fmt.Printf("%d: %s\r\n", i, item.ClassName())
		}
		return false
	} else if (len(os.Args) == 2 && os.Args[1] == "-i") ||
		(len(os.Args) == 4 && os.Args[1] == "-i") {
		//获取用户选择
		for i, item := range classOptions {
			fmt.Printf("%d: %s\r\n", i, item.ClassName())
		}
		fmt.Print("Index:")
		fmt.Scanf("%d", &nIndexChoice)

		//检查范围
		if nIndexChoice < 0 || nIndexChoice > len(classOptions)-1 {
			fmt.Println("Index out of range.")
			return false
		}

		if len(os.Args) == 4 {
			dateBegin, _ = time.ParseInLocation("200601021504", os.Args[2], time.Local)
			dateEnd, _ = time.ParseInLocation("200601021504", os.Args[3], time.Local)
		}
		//选择设置
		MailFetchConfig = classOptions[nIndexChoice]
	} else {
		strUsage := strings.Replace(usageTemplate, "<py>", filepath.Base(os.Args[0]), -1)
		fmt.Printf(strUsage, dateBegin.Format("200601021504"), dateEnd.Format("200601021504"),
			dateBegin.Format("200601021504"), dateEnd.Format("200601021504"))
		return false
	}

	fmt.Println("设置信息获取完毕，开始下载邮件...")
	//用户设置时间覆盖下载时间设置
	MailFetchConfig.DateStart, MailFetchConfig.DateEnd =
		dateBegin, dateEnd
	return true
}
func main() {

	if setConfigs() {
		Run()
	}
}
