package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mailgun/mailgun-go/v4"
)

const urlToMonitor = "https://sites.google.com/view/bigdatafelician/website-monitoring"

const emailSender = "filippo@mg.neysofu.me"
const emailRecipient = "filippo@neysofu.me"
const mailgunDomain = "mg.neysofu.me"

const cooloffPeriod = 60 * time.Second

func main() {
	// Pretend we're running a web service so render.com doesn't kill us
	port := os.Getenv("PORT")
	go http.ListenAndServe(":"+port, nil)

	mailgunApiKey := os.Getenv("MAILGUN_API_KEY")
	fmt.Println("Mailgun API Key: ", mailgunApiKey[:4], "...")
	numSeenNames := 3

	for {
		names, _ := findPeopleNames(urlToMonitor)
		fmt.Println("Found", len(names), "names")

		for _, name := range names[numSeenNames:] {
			fmt.Println("Found a new name: ", name)
			go sendEmail(mailgunApiKey, name)
		}
		numSeenNames = len(names)

		time.Sleep(cooloffPeriod)
	}
}

func findPeopleNames(url string) ([]string, error) {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return []string{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return []string{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return []string{}, err
	}

	names := []string{}
	doc.Find("ul li p").Each(func(i int, s *goquery.Selection) {
		names = append(names, s.Text())
	})

	return names, nil
}

func sendEmail(mailgunApiKey string, subject string) {
	mg := mailgun.NewMailgun(mailgunDomain, mailgunApiKey)
	body := "Hello from Mailgun Go!"

	// The message object allows you to add attachments and Bcc recipients
	message := mg.NewMessage(emailSender, subject, body, emailRecipient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Send the message with a 10 second timeout
	resp, id, err := mg.Send(ctx, message)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ID: %s Resp: %s\n", id, resp)
}
