package main

import (
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

//MailFetchConfig refered to config txt file
var MailFetchConfig TagClassInfo

func downloadAttach(c *client.Client, downloadSet *imap.SeqSet) {
	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}
	log.Println("保存至: ", MailFetchConfig.rootPath)
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

				fileBytes, _ := ioutil.ReadAll(p.Body)
				file, err := os.Create(path.Join(MailFetchConfig.rootPath, filename))
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				// Save file
				file.Write(fileBytes)
				log.Println("已保存:", filename)

				// Remove name form violating list
				splits := strings.Split(filename, MailFetchConfig.delimiter)
				if len(splits) == 3 {
					removeName(splits[1])
				}
				break // Only download first attchment
			}
		}
	}

}

func isMailSatisfied(msg *imap.Message) bool {
	strSubject := strings.ToUpper(decodeMailSubject(msg.Envelope.Subject))
	mailTime := msg.Envelope.Date

	if isNameCorrect(strSubject, MailFetchConfig.prefixFlag, MailFetchConfig.delimiter) != true {
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

// Returns a set contains all mails avaliable
func getMailsSet(client *client.Client) (set *imap.SeqSet) {
	seqset := new(imap.SeqSet)
	log.Println("Messages count: ", client.Mailbox().Messages)
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

// Return a set contains all mails meeting the time and name requirements
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

func createRootPath() {
	// Create the root path for storing the attchments
	dstPath := path.Join(MailFetchConfig.rootPath, MailFetchConfig.prefixFlag, MailFetchConfig.prefixFlag+"_"+time.Now().Format("20060102"))
	os.MkdirAll(dstPath, 0777)
	MailFetchConfig.rootPath = dstPath
	log.Println("存储路径:", MailFetchConfig.rootPath)
}

func connect2Server() *client.Client {
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
	return c
}

func downloadMails() {
	c := connect2Server()
	// Get all avaliable messages by header info
	seqset := getMailsSet(c)
	downloadset := getSatisfiedMails(c, seqset)

	// Download attchment whose mail is OK
	log.Println("将要下载:", downloadset)
	downloadAttach(c, downloadset)
}

//Run starts downloading mails' attachement and classify
func Run() {
	createRootPath()

	downloadMails()

	recordLogs()
}
