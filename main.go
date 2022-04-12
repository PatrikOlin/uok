package main

import (
	"bufio"
	"encoding/gob"
	"errors"
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
var stateGob = ".uok_state.gob"

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
	var down []string
	for key, value := range res {
		if strings.Contains(value, "200") {
			fmt.Println(style.Bold(false).Render(status(value) + "   " + key))
		} else {
			fmt.Println(style.Bold(true).Render(issueStatus(value) + "   " + key))

			down = append(down, fmt.Sprintf("%s returned %s", key, value))
		}
	}

	if token != "" {
		prevState, err := load()
		if errors.Is(err, os.ErrNotExist) {
			notify(down)
		} else if err != nil {
			log.Fatalln(err)
		} else {
			filtered := filter(down, prevState)
			fmt.Println("failed without duplicates, ", filtered)

			if len(filtered) > 0 {
				notify(filtered)
			}
		}
	}

	save(down)
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

func save(slice []string) {
	m := make(map[string]string)
	for _, x := range slice {
		x := strings.Split(x, " ")[0]
		m[x] = x
	}

	file, err := os.Create(stateGob)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	encoder.Encode(m)
}

func load() (map[string]string, error) {
	var data map[string]string

	file, err := os.Open(stateGob)
	if err != nil {
		return data, err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&data)

	if err != nil {
		return data, err
	}

	return data, nil
}

func filter(s []string, m map[string]string) []string {
	var filtered []string
	for i, val := range s {
		x := strings.Split(val, " ")[0]
		if _, ok := m[x]; !ok {
			filtered = append(filtered, s[i])
		}

	}
	return filtered
}
