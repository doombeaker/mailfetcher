package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime/quotedprintable"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

//SplitSubject splits subject in RFC form to charset, encode type and subject content
func splitSubject(strSubject string) (strCharset string, encodeType string, strContent string) {
	strTest := strSubject
	splits := strings.SplitN(strTest[2:], "?", 2)

	strCharset = strings.ToUpper(splits[0])

	splits = strings.SplitN(splits[1], "?", 2)
	encodeType = strings.ToUpper(splits[0])

	strContent = splits[1][:len(splits[1])-2]
	return
}

//isNameCorrect checks title's form. eg: NR302_yaochi_20190528CPointer.rar
func isNameCorrect(titleName string, strPrefix string) bool {
	splits := strings.Split(titleName, "_")

	if strings.HasPrefix(titleName, strPrefix) && len(splits) == 3 {
		return true
	}

	return false
}

func decodeMailSubject(subject string) string {
	var strRet string

	if strings.HasPrefix(subject, "=?") {
		splits := strings.Split(subject, " ")

		for _, item := range splits {
			strTemp := decodeRFCString(item)
			strRet += strTemp
		}
	} else {
		strRet = subject
	}

	return strRet
}

//decodeRFCString decode strings like:
//=?gb18030?B?TlIxMDIzX7PCx65fMjAxOTA1MjU=?=
//=?utf-8?Q?NR1023=5F=E9=99=88=E9=B8=BF=E6=AF=85=5F20190525?=
//3-parts: charset encoding encoded-text
//=?charset?encoding?encoded-text?=
//charset:
//encoding: B or Q, B meaings base64 Q means quoted-printable
func decodeRFCString(subject string) string {

	var dataBytes []byte
	var strRet string
	var err error

	subCharset, encodeType, content := splitSubject(subject)

	//decode by base64 or quoted-print
	switch encodeType {
	case "B":
		//decode by base64
		data, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			log.Fatal(err)
		}

		dataBytes = data
	case "Q":
		data, err := ioutil.ReadAll(quotedprintable.NewReader(strings.NewReader(content)))

		if err != nil {
			log.Fatal(err)
		}

		dataBytes = data
	}

	//convert content via charaset
	// convert by GB18030
	if strings.ToUpper(subCharset) == "GB18030" {
		decodeBytes, _ := simplifiedchinese.GB18030.NewDecoder().Bytes(dataBytes)
		strRet = string(decodeBytes)
	} else if strings.ToUpper(subCharset) == "UTF-8" {
		strRet = string(dataBytes)
	} else if strings.ToUpper(subCharset) == "GBK" {
		decodeBytes, _ := simplifiedchinese.GBK.NewDecoder().Bytes(dataBytes)
		strRet = string(decodeBytes)
	}

	if err != nil {
		fmt.Println("error:", err)
		return ""
	}

	return strRet
}
