package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/client"
	"github.com/geziyor/geziyor/export"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	xbook_1()
	urls := readIndex()
	xbook(urls)
	merge()
}
func xbook_1() {
	geziyor.NewGeziyor(&geziyor.Options{
		StartURLs: []string{"https://book.xbookcn.net/search/label/%E5%B0%91%E5%B9%B4%E9%98%BF%E5%AE%BE"},
		ParseFunc: parseIndex,
		Exporters: []export.Exporter{&export.JSON{
			FileName: "results/index.json",
		}},
	}).Start()
}

func parseIndex(g *geziyor.Geziyor, r *client.Response) {
	r.HTMLDoc.Find("h3").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Find("a").Attr("href")
		// Export Data
		g.Exports <- map[string]interface{}{
			"number": i,
			"title":  s.Find("a").Text(),
			"url":    url,
		}
	})

}
func xbook(urls []string) {
	if len(urls) == 0 {
		return
	}
	for k, url := range urls {
		geziyor.NewGeziyor(&geziyor.Options{
			StartURLs: []string{url},
			ParseFunc: parseArticle,
			Exporters: []export.Exporter{&export.CSV{
				FileName: "results/article_" + strconv.Itoa(k+1) + ".csv",
			}},
		}).Start()
	}

}
func parseArticle(g *geziyor.Geziyor, r *client.Response) {
	title := r.HTMLDoc.Find("div.post-outer").Find("h3").Text()
	g.Exports <- map[string]interface{}{
		"title": title,
	}
	r.HTMLDoc.Find("div.post-outer").Find("div.post-body > p").Each(func(i int, s *goquery.Selection) {
		content := strings.Trim(s.Text(), " ")
		if len(content) > 0 {
			g.Exports <- map[string]interface{}{
				"content": s.Text(),
			}
		}

	})
}

func readIndex() []string {
	fileContent, err := os.OpenFile("results/index.json", os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	defer fileContent.Close()

	var res []map[string]interface{}
	decoder := json.NewDecoder(fileContent)
	err = decoder.Decode(&res)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	var urls []string
	for _, v := range res {
		urls = append(urls, v["url"].(string))
	}
	return urls
}

// removeSpaces 去掉字符串中的空格
func removeSpaces(s string) string {
	str := strings.Trim(s, "\t")
	str = strings.ReplaceAll(str, "\"", "")
	return strings.ReplaceAll(str, " ", "")
}

// processCSVFile 读取CSV文件，处理每一行数据的空格，并保存为TXT文件
func processCSVFile(filePath string) error {
	// 打开CSV文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建或清空对应的TXT文件
	txtFilePath := strings.TrimSuffix(filePath, ".csv") + ".txt"
	err = os.WriteFile(txtFilePath, []byte{}, 0644)
	if err != nil {
		return err
	}

	// 使用bufio读取CSV文件每一行
	scanner := bufio.NewScanner(file)
	writer, err := os.OpenFile(txtFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer writer.Close()

	for scanner.Scan() {
		line := scanner.Text()
		processedLine := removeSpaces(line) + "\n"
		if _, err := writer.WriteString(processedLine); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	fmt.Printf("Processed and saved as %s\n", txtFilePath)
	return nil
}

// traverseAndProcess 目录遍历并处理其中的CSV文件
func traverseAndProcess(dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasPrefix(info.Name(), "article_") && strings.HasSuffix(info.Name(), ".csv") {
			fmt.Printf("Processing %s...\n", path)
			return processCSVFile(path)
		}
		return nil
	})
}

func merge() {
	dirPath := "./results" // 请根据实际情况修改为你的目录路径
	fmt.Println("Starting processing...")
	if err := traverseAndProcess(dirPath); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Processing completed.")
	}
}
