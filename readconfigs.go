package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

//TagClassInfo 存储配置信息
type TagClassInfo struct {
	className    string
	homeworkPath string
	mailserver   string
	mailUser     string
	mailPassword string
	prefixFlag   string
	stuLists     []string
	DateStart    time.Time
	DateEnd      time.Time
	VIOLATELIST  []string
}

//ClassName 班级名称
func (clsInfo *TagClassInfo) ClassName() string {
	return clsInfo.className
}

// readConfig Read config options form txt file
func readConfig(txtPath string, classConfigs *[]TagClassInfo) {
	var currentClassInfo TagClassInfo

	//fmt.Println("read configs from: ", txtPath)

	file, err := os.Open(txtPath)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	buf := bufio.NewReader(file)

	//读取配置信息
LabelReadOptions:
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			break LabelReadOptions
		}

		//解析配置key和value
		keyvalue := strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		key := strings.SplitN(keyvalue, "=", 2)[0]
		value := strings.SplitN(keyvalue, "=", 2)[1]

		switch key {
		case "homework_path":
			currentClassInfo.homeworkPath = value
		case "mailserver":
			currentClassInfo.mailserver = value
		case "mail_user":
			currentClassInfo.mailUser = value
		case "mail_passwd":
			currentClassInfo.mailPassword = value
		case "prefix_flag":
			currentClassInfo.prefixFlag = value
			currentClassInfo.className = value
		default:
			fmt.Println("Unknown: ", key, value)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

LabelStudents:
	//读取学员名单信息
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			break LabelStudents
		}

		//初始化学员列表和违纪列表
		currentClassInfo.stuLists = append(currentClassInfo.stuLists, line)
		currentClassInfo.VIOLATELIST = append(currentClassInfo.VIOLATELIST, line)

		if err != nil {
			if err == io.EOF {
				break LabelStudents
			}
			log.Fatal(err)
		}
	}

	*classConfigs = append(*classConfigs, currentClassInfo)
	//fmt.Println(txtPath, "done.")
}

//ReadConfigDir Read configs from config txt files
func ReadConfigDir(configpath string) []TagClassInfo {
	var configTxtFiles []string
	var classConfigs []TagClassInfo

	if configpath == "" {
		configpath = "./"
	}

	files, err := ioutil.ReadDir(configpath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "txt") {
			configTxtFiles = append(configTxtFiles, file.Name())
		}
	}

	for _, item := range configTxtFiles {
		readConfig(path.Join(configpath, item), &classConfigs)
	}

	return classConfigs
}
