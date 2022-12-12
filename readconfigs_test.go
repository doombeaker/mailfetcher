package main

import (
	"testing"
)

func TestReadConfig(t *testing.T) {
	var classConfigs []TagClassInfo
	readConfig("./configs/EXAMPLE.txt", &classConfigs)
	config := classConfigs[0]

	if config.className != "SS1201" {
		t.Errorf("config.className expected 'SS1201', but %s got", config.className)
	}

	if config.delimiter != "-" {
		t.Errorf("config.delimiter expected '-', but %s got", config.delimiter)
	}

	if config.homeworkPath != "c:\\homework" {
		t.Errorf("config.homeworkPath expected 'c:\\homework', but %s got", config.homeworkPath)
	}

	if config.mailUser != "username" {
		t.Errorf("config.mailUser expected 'username', but %s got", config.mailUser)
	}

	if config.mailPassword != "password" {
		t.Errorf("config.mailUser expected 'password', but %s got", config.mailPassword)
	}

	if config.MAXMAILS != 40 {
		t.Errorf("config.MAXMAILS expected 40, but %d got", config.MAXMAILS)
	}

	stu_list := config.stuLists
	if stu_list[0] != "张三" && stu_list[1] != "李四" && stu_list[2] != "王五" {
		t.Errorf("config.stuLists expected [张三, 李四, 王五], but [%s, %s, %s] got",
			stu_list[0], stu_list[1], stu_list[2])
	}
}
