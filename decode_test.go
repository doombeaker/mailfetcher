package main

import (
	"testing"
)

func TestMailCheckUtils(t *testing.T) {
	if ans := isNameCorrect("SS1201-张三-单元测试", "SS1201", "-"); ans != true {
		t.Errorf(`Failed at isNameCorrect("SS1201-张三-单元测试", "SS1201", "-")`)
	}

	if ans := isNameCorrect("SS1201_张三_单元测试", "SS1201", "_"); ans != true {
		t.Errorf(`Failed at isNameCorrect(SS1201_张三_单元测试, "SS1201", "_")`)
	}
}

func TestDecodeSubject(t *testing.T) {
	test_io := make(map[string]string)
	test_io["=?gb18030?B?TlIxMDIzX7PCx65fMjAxOTA1MjU=?="] = "NR1023_陈钱_20190525"
	test_io["=?utf-8?Q?NR1023=5F=E9=99=88=E9=B8=BF=E6=AF=85=5F20190525?="] = "NR1023_陈鸿毅_20190525"
	test_io["=??B?aGVsbG93b3JsZA==?="] = "helloworld"
	for k := range test_io {
		if ans := decodeMailSubject(k); ans != test_io[k] {
			t.Errorf("%s expected, but %s got", test_io[k], ans)
		}
	}
}
