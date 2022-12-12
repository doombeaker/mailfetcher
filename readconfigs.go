package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
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
	MAXMAILS     uint32
	delimiter    string
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

	file, err := os.Open(txtPath)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	buf := bufio.NewReader(file)

LabelReadOptions:
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			break LabelReadOptions
		}

		// Parse key and value
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
		case "maxmail":
			maxmails, _ := strconv.ParseInt(value, 10, 32)
			currentClassInfo.MAXMAILS = uint32(maxmails)
		case "delimiter":
			currentClassInfo.delimiter = value
		default:
			fmt.Println("Unknown: ", key, value)
		}

		if err != nil {
			log.Fatal(err)
		}
	}

LabelStudents:
	// read name line by line
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)

		if line == "" {
			break LabelStudents
		}

		// Initialize name lists and violated lists
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
