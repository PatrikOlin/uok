package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/pflag"
)

var token string
var (
	subtle  = lipgloss.Color("#ebe8e8")
	special = lipgloss.Color("#73F59F")

	style = lipgloss.NewStyle().
		Foreground(subtle)

	status      = lipgloss.NewStyle().Foreground(special).Render
	issueStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Blink(true).Render
)

func init() {
	pflag.StringVarP(&token, "gotify token", "t", "", "the gotify token used for sending notifications to gotify")
	pflag.Parse()
}

func main() {
	urls, err := getURLs()
	if err != nil {
		log.Fatalln("Could not read uok-urls file", err)
	}

	res := make(map[string]string)
	for _, url := range urls {
		status := makeRequest(url)

		res[url] = status
	}

	printBanner()
	var failed []string
	for key, value := range res {
		if strings.Contains(value, "200") {
			fmt.Println(style.Bold(false).Render(status(value) + "   " + key))
		} else {
			fmt.Println(style.Bold(true).Render(issueStatus(value) + "   " + key))

			failed = append(failed, fmt.Sprintf("%s returned %s", key, value))
		}
	}

	if token != "" {
		notify(failed)
	}
}

func getURLs() ([]string, error) {
	f, err := os.Open("uok-urls")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var urls []string

	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	return urls, scanner.Err()
}

func makeRequest(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return err.Error()
	}

	return resp.Status
}

func notify(msgs []string) {
	msg := strings.Join(msgs, "\n")
	u := fmt.Sprintf("https://notify.olin.dev/message?token=%s", token)
	http.PostForm(u, url.Values{"message": {msg}, "title": {"Houston, we have a problem"}})
	fmt.Println(style.Bold(false).Render("Sent notifications to gotify"))
}

func printBanner() {
	fig := figure.NewFigure("u ok?", "slant", true)
	fig.Print()
	fmt.Println()
}
