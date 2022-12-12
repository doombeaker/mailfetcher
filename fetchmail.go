package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

//MailFetchConfig 包含了下载配置信息的结构体遍历
var MailFetchConfig TagClassInfo

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

func downloadAttach(c *client.Client, downloadSet *imap.SeqSet) {
	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}
	log.Println("保存至: ", MailFetchConfig.homeworkPath)
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(downloadSet, items, messages)
	}()
	for msg := range messages {

		r := msg.GetBody(section)
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		// Read each mail's part
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			switch h := p.Header.(type) {
			case *mail.AttachmentHeader:
				filename, _ := h.Filename()

				//保存作业
				fileBytes, _ := ioutil.ReadAll(p.Body)
				file, err := os.Create(path.Join(MailFetchConfig.homeworkPath, filename))
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				//data, _ := base64.StdEncoding.DecodeString(string(fileBytes))
				file.Write(fileBytes)
				log.Println("已保存:", filename)

				//移除违纪名单
				splits := strings.Split(filename, MailFetchConfig.delimiter)
				if len(splits) == 3 {
					removeStuName(splits[1])
				}
			}
		}
	}

}

func isMailSatisfied(msg *imap.Message) bool {
	strSubject := strings.ToUpper(decodeMailSubject(msg.Envelope.Subject))
	mailTime := msg.Envelope.Date

	if isNameCorrect(strSubject, MailFetchConfig.prefixFlag) != true {
		log.Printf("%s: Name Wrong (%s)\r\n", strSubject, MailFetchConfig.prefixFlag)
		return false
	}
	if mailTime.After(MailFetchConfig.DateEnd) || mailTime.Before(MailFetchConfig.DateStart) {
		log.Printf("%s: Time Wrong (%s %s-%s)\r\n", strSubject, mailTime.Format("200601021504"),
			MailFetchConfig.DateStart.Format("200601021504"),
			MailFetchConfig.DateEnd.Format("200601021504"))
		return false
	}

	log.Println(strSubject, "OK")
	return true
}

//GetMailsSet returns a set contains UID of mails matched requirments
func getMailsSet(client *client.Client) (set *imap.SeqSet) {
	seqset := new(imap.SeqSet)
	log.Println("Messages count: ", client.Mailbox().Messages)
	// Get the last 4 messages
	mbox := client.Mailbox()
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > MailFetchConfig.MAXMAILS {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - MailFetchConfig.MAXMAILS
	}
	seqset.AddRange(from, to)
	return seqset
}

func getSatisfiedMails(c *client.Client, seqset *imap.SeqSet) *imap.SeqSet {
	items := []imap.FetchItem{imap.FetchEnvelope}
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	downloadset := new(imap.SeqSet)
	for msg := range messages {
		if isMailSatisfied(msg) {
			downloadset.AddNum(msg.SeqNum)
		}
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	return downloadset
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

func saveViolateStudents() {
	saveViolates2Txt()
	saveViolates2Sqlite()
}

func createHomeworkPath() {
	//创建存储路径
	dstPath := path.Join(MailFetchConfig.homeworkPath, MailFetchConfig.prefixFlag, MailFetchConfig.prefixFlag+"_"+time.Now().Format("20060102"))
	os.MkdirAll(dstPath, 0777)
	MailFetchConfig.homeworkPath = dstPath
	log.Println("存储路径:", MailFetchConfig.homeworkPath)
}

func fetchToSaveMails() {
	// Connect to the server
	c, err := client.DialTLS(MailFetchConfig.mailserver, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to: ", MailFetchConfig.mailserver)

	// Authenticate
	if err := c.Login(MailFetchConfig.mailUser, MailFetchConfig.mailPassword); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in: ", MailFetchConfig.mailUser)

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Get the last messages
	seqset := getMailsSet(c)

	downloadset := getSatisfiedMails(c, seqset)

	log.Println("将要下载:", downloadset)
	downloadAttach(c, downloadset)
}

//Run starts downloading mails' attachement and classify
func Run() {
	createHomeworkPath()

	fetchToSaveMails()

	saveViolateStudents()
}
