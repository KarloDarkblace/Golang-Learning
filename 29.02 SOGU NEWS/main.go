package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func createFile() *os.File {
	file, err := os.Create("news.txt")

	if err != nil {
		log.Fatalf("Ошибка при создании файла: %v", err)
	}

	return file
}

func startParsing() []string {
	var titles []string
	for page := 1; ; page++ {
		pageTitles, err := fetchNewsTitlesForPage(page)
		if len(pageTitles) == 0 || err != nil {
			fmt.Printf("Новости закончились на странице %d\n", page)
			break
		}
		titles = append(titles, pageTitles...)
	}
	return titles
}

func printToFile(file *os.File, titles []string) {
	for i, title := range titles {
		_, err := file.WriteString(strconv.Itoa(i+1) + ": " + title + "\n")
		if err != nil {
			log.Fatal("Ошибка при записи в файл: ", err)
		}
	}
}

func fetchNewsTitlesForPage(page int) ([]string, error) {
	var titles []string
	data := url.Values{}
	data.Set("paged", strconv.Itoa(page))

	req, err := http.NewRequest("POST", "https://www.nosu.ru/category/news/", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("ERROR | #1 | END OF PAGES")
		} else {
			log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".content-block.content-text.news-item .title a").Each(func(i int, s *goquery.Selection) {
		titles = append(titles, s.Text())
	})

	return titles, nil
}

func main() {
	file := createFile()
	defer file.Close()

	titles := startParsing()
	printToFile(file, titles)

	fmt.Println("GOOD ENDING")
}
