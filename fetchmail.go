package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path"
	"strings"
	"time"

	"github.com/mxk/go-imap/imap"
)

//MailFetchConfig 包含了下载配置信息的结构体遍历
var MailFetchConfig TagClassInfo

//MAXMAILS 表示遍历邮件的最大数目
var MAXMAILS uint32 = 40

//getTimeFromDateString 解析mail中日期字符串得到time对象
func getTimeFromDateString(strDate string) time.Time {

	nIndexBracket := strings.Index(strDate, "(")
	if nIndexBracket != -1 {
		strDate = strDate[:nIndexBracket-1]
	}
	//获取本地location
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation("Mon, _2 Jan 2006 15:04:05 -0700", strDate, loc)

	return theTime
}

//RemoveStuName Remove Student's Name from VIOLATELIST
func removeStuName(stuName string) {
	for i, item := range MailFetchConfig.VIOLATELIST {
		if item == stuName {
			MailFetchConfig.VIOLATELIST = append(MailFetchConfig.VIOLATELIST[:i],
				MailFetchConfig.VIOLATELIST[i+1:]...)
			return
		}
	}
}

//SaveViolates2DB Saves violated records
func saveViolates2Sqlite() {

	//违纪学生存数据库
	db, err := sql.Open("sqlite3", "./data.db")
	defer db.Close()

	if err != nil {
		log.Println(err)
	}

	for _, item := range MailFetchConfig.VIOLATELIST {
		stmt, err := db.Prepare(`INSERT INTO violate (clsname, stuname, date) VALUES (?, ?, datetime('now', 'localtime'))`)
		if err != nil {
			log.Println(err)
		} else {
			_, err := stmt.Exec(MailFetchConfig.className, item)
			if err != nil {
				log.Println(err)
			}
		}

	}
}

func myPraseAttachMent(part *multipart.Part) {
	myMediaType, myParams, err := mime.ParseMediaType(part.Header.Get("Content-Disposition"))

	if err != nil {
		log.Fatal(err)
		return
	}
	filename := decodeMailSubject(myParams["filename"])
	enocdeType := part.Header.Get("Content-Transfer-Encoding")

	//保存作业
	fileBytes, _ := ioutil.ReadAll(part)
	file, err := os.Create(path.Join(MailFetchConfig.homeworkPath, filename))
	fmt.Println(path.Join(MailFetchConfig.homeworkPath, filename))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	data, _ := base64.StdEncoding.DecodeString(string(fileBytes))
	file.Write(data)
	fmt.Println("已保存:", myMediaType, filename, enocdeType, "\r\n")

	splits := strings.Split(filename, "_")
	if len(splits) == 3 {
		removeStuName(splits[1])
	}

}

func myParseMailMsg(msg *mail.Message) {
	header := msg.Header

	mediaType, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		log.Fatal(err)
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])

		for i := 0; ; i++ {
			part, err := mr.NextPart()
			//fmt.Println("--------Multi-Part:", i, "-------")
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Fatal(err)
			}

			//只下一个附件
			if part.Header.Get("Content-Disposition") != "" {
				myPraseAttachMent(part)
				return
			}
		}
	}
}

func downloadAttach(seqSet *imap.SeqSet, cmd *imap.Command, client *imap.Client) {
	// Fetch the headers of the 3 most recent messages
	cmd, _ = client.Fetch(seqSet, "BODY[]")

	// Process responses while the command is running
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		client.Recv(-1)

		// Process command data
		for _, rsp := range cmd.Data {
			msgIno := rsp.MessageInfo()
			msgBytes := imap.AsBytes(msgIno.Attrs["BODY[]"])

			if msg, err := mail.ReadMessage(bytes.NewReader(msgBytes)); msg != nil {
				if err != nil {
					log.Fatal(err)
				}

				myParseMailMsg(msg)
			}
		}
		cmd.Data = nil

		// Process unilateral server data
		for _, rsp := range cmd.Data {
			fmt.Println("Server data:", rsp)
		}
		cmd.Data = nil
	}
}
func isMailSatisfied(msgHeader *mail.Header) bool {
	strSubject := strings.ToUpper(decodeMailSubject(msgHeader.Get("Subject")))

	strDate := msgHeader.Get("Date")
	mailTime := getTimeFromDateString(strDate)

	fmt.Printf("%s: ", strSubject)
	if isNameCorrect(strSubject, MailFetchConfig.prefixFlag) != true {
		fmt.Printf("Name Wrong (%s)\r\n", MailFetchConfig.prefixFlag)
		return false
	}

	if mailTime.After(MailFetchConfig.DateEnd) || mailTime.Before(MailFetchConfig.DateStart) {
		fmt.Printf("Time Wrong (%s %s-%s)\r\n", mailTime.Format("200601021504"),
			MailFetchConfig.DateStart.Format("200601021504"),
			MailFetchConfig.DateEnd.Format("200601021504"))
		return false
	}

	fmt.Println("OK")
	return true

}

//GetMailsSet returns a set contains UID of mails matched requirments
func getMailsSet(client *imap.Client) (set *imap.SeqSet, err error) {
	// Fetch the headers of the 3 most recent messages
	set, err = imap.NewSeqSet("")
	if client.Mailbox.Messages > MAXMAILS {
		set.AddRange(client.Mailbox.Messages-MAXMAILS, client.Mailbox.Messages)
	} else {
		set.Add("1:*")
	}

	return set, err
}

func getSatisfiedMails(set *imap.SeqSet, cmd *imap.Command, client *imap.Client) *imap.SeqSet {
	cmd, _ = client.Fetch(set, "RFC822.HEADER")
	var seqSet imap.SeqSet

	// Process responses while the command is running
	fmt.Println("\n遍历邮件中...")
	for cmd.InProgress() {
		// Wait for the next response (no timeout)
		client.Recv(-1)

		// Process command data
		for _, rsp := range cmd.Data {
			mailInfo := rsp.MessageInfo()
			header := imap.AsBytes(mailInfo.Attrs["RFC822.HEADER"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				if isMailSatisfied(&msg.Header) {
					seqSet.AddNum(mailInfo.Seq)
				}
			}
		}
		cmd.Data = nil

		// Process unilateral server data
		for _, rsp := range cmd.Data {
			fmt.Println("Server data:", rsp)
		}
		cmd.Data = nil
	}

	return &seqSet
}

func saveViolates2Txt() {
	//打印违纪名单
	outputTemplate := `
	<class>    <date>
	应交:%d		实交:%d


	班级名单:
	%s


	违纪名单:
	%s
	`
	outputTemplate = strings.Replace(outputTemplate, "<class>", MailFetchConfig.className, 1)
	outputTemplate = strings.Replace(outputTemplate, "<date>", time.Now().Format(time.RFC1123Z), 1)
	strAll := strings.Join(MailFetchConfig.stuLists, "    ")
	strViolate := strings.Join(MailFetchConfig.VIOLATELIST, "    ")

	outputText := fmt.Sprintf(outputTemplate, len(MailFetchConfig.stuLists),
		len(MailFetchConfig.stuLists)-len(MailFetchConfig.VIOLATELIST),
		strAll, strViolate)
	fmt.Print(outputText)

	file, _ := os.Create(path.Join(MailFetchConfig.homeworkPath, "违纪统计.txt"))
	defer file.Close()

	io.WriteString(file, outputText)
}

func saveViolatStus() {
	saveViolates2Txt()
	saveViolates2Sqlite()
}

func createHomeworkPath() {
	//创建存储路径
	dstPath := path.Join(MailFetchConfig.homeworkPath, MailFetchConfig.prefixFlag, MailFetchConfig.prefixFlag+"_"+time.Now().Format("20060102"))
	os.MkdirAll(dstPath, 0777)
	MailFetchConfig.homeworkPath = dstPath
	fmt.Println("存储路径:", MailFetchConfig.homeworkPath)
}

func fetchToSaveMails() {
	//
	// Note: most of error handling code is omitted for brevity
	//
	var (
		c   *imap.Client
		cmd *imap.Command
		rsp *imap.Response
	)

	// Connect to the server
	c, _ = imap.Dial(MailFetchConfig.mailserver)

	// Remember to log out and close the connection when finished
	defer c.Logout(30 * time.Second)

	// Print server greeting (first response in the unilateral server data queue)
	fmt.Println("Server says hello:", c.Data[0].Info)
	c.Data = nil

	// Enable encryption, if supported by the server
	if c.Caps["STARTTLS"] {
		c.StartTLS(nil)
	}

	// Authenticate
	if c.State() == imap.Login {
		c.Login(MailFetchConfig.mailUser, MailFetchConfig.mailPassword)
	}

	//List all top-level mailboxes, wait for the command to finish
	cmd, _ = imap.Wait(c.List("", "%"))

	// Check for new unilateral server data responses
	for _, rsp = range c.Data {
		fmt.Println("Server data:", rsp)
	}
	c.Data = nil

	// Open a mailbox (synchronous command - no need for imap.Wait)
	c.Select("INBOX", true)

	// Fetch the headers of the 3 most recent messages
	set, _ := getMailsSet(c)

	setFetchMail := getSatisfiedMails(set, cmd, c)

	fmt.Println("将要下载:", setFetchMail)
	downloadAttach(setFetchMail, cmd, c)

	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			fmt.Println("Fetch command aborted")
		} else {
			fmt.Println("Fetch error:", rsp.Info)
		}
	}
}

//Run starts downloading mails' attachement and classify
func Run() {
	createHomeworkPath()

	fetchToSaveMails()

	saveViolatStus()
}
