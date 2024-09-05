package lib

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

func Start() {
	help := flag.Bool("h", false, "帮助菜单")
	threads := flag.Int("t", 8, "设置线程数（默认为 8），如果指定 -u，将被忽略")
	domain := flag.String("u", "", "单个域名地址")
	file := flag.String("f", "", "包含多个域名地址的文件")

	flag.Parse()

	if *help {
		fmt.Println("Such:main -u baidu.com")
		fmt.Println("Such:main -f domain.txt -t 8")
		fmt.Println("Usage:")
		fmt.Println("      -h         Help menu.")
		fmt.Println("      -t <num>   thread,Default to 8 threads.")
		fmt.Println("      -u <url>   domain")
		fmt.Println("      -f <file>  file")
		return
	}

	if *domain != "" && *file != "" {
		fmt.Println(Red("[-] 不能同时使用 -u 和 -f 选项."))
		return
	}

	if *domain != "" {
		*threads = 1
	}

	if *domain == "" && *file == "" {
		fmt.Println(Red("[-] 请提供域名地址（-u）或文件（-f）"))
		return
	}

	var domains []string
	re := regexp.MustCompile(`(?:https?:\/\/)?(?:www\.)?([^\/:]+)`)

	if *file != "" {
		fileContent, err := ioutil.ReadFile(*file)
		if err != nil {
			log.Fatalf("[-] 读取文件出错: %v", err)
		}

		lines := strings.Split(string(fileContent), "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" {
				matches := re.FindStringSubmatch(trimmedLine)
				if len(matches) > 1 {
					domains = append(domains, matches[1])
				}
			}
		}
	} else if *domain != "" {
		trimmedDomain := strings.TrimSpace(*domain)
		if trimmedDomain != "" {
			matches := re.FindStringSubmatch(trimmedDomain)
			if len(matches) > 1 {
				domains = append(domains, matches[1])
			}
		}
	}

	currentTime := time.Now().Format("2006-01-02T15-04-05")
	folderName := fmt.Sprintf("result-%s", currentTime)
	err := os.MkdirAll(folderName, os.ModePerm)
	if err != nil {
		log.Fatalf("[-] 创建文件夹 %s 失败: %v", folderName, err)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, *threads) // 控制并发数

	allFile, err := os.OpenFile(fmt.Sprintf("%s/all.txt", folderName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("[-] 打开 %s/all.txt 失败: %v", folderName, err)
	}
	defer allFile.Close()

	jsFile, err := os.OpenFile(fmt.Sprintf("%s/js.txt", folderName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("[-] 打开 %s/js.txt 失败: %v", folderName, err)
	}
	defer jsFile.Close()

	paramsFile, err := os.OpenFile(fmt.Sprintf("%s/parameters.txt", folderName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("[-] 打开 %s/parameters.txt 失败: %v", folderName, err)
	}
	defer paramsFile.Close()

	xlsFile, err := os.OpenFile(fmt.Sprintf("%s/xls.txt", folderName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("[-] 打开 %s/xls.txt 失败: %v", folderName, err)
	}
	defer xlsFile.Close()

	csvFile, err := os.OpenFile(fmt.Sprintf("%s/csv.txt", folderName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("[-] 打开 %s/csv.txt 失败: %v", folderName, err)
	}
	defer csvFile.Close()

	pdfFile, err := os.OpenFile(fmt.Sprintf("%s/pdf.txt", folderName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("[-] 打开 %s/pdf.txt 失败: %v", folderName, err)
	}
	defer pdfFile.Close()

	jsRegex := regexp.MustCompile(`(?i)\.js`)
	paramsRegex := regexp.MustCompile(`(?i)\?`)
	xlsRegex := regexp.MustCompile(`(?i)\.xls`)
	csvRegex := regexp.MustCompile(`(?i)\.csv`)
	pdfRegex := regexp.MustCompile(`(?i)\.pdf`)

	for _, domain := range domains {
		sem <- struct{}{}
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			defer func() { <-sem }()

			url := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=*.%s&output=text&fl=original&collapse=urlkey&from=", domain)
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("[-] 请求 %s 失败: %v", url, err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("[-] 读取响应内容失败: %v", err)
				return
			}
			lines := strings.Split(string(body), "\n")

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				if _, err := allFile.WriteString(line + "\n"); err != nil {
					log.Printf("[-] 写入 %s/all.txt 失败: %v", folderName, err)
				}

				if jsRegex.MatchString(line) {
					fmt.Println(White("[+] Find js：" + line))
					if _, err := jsFile.WriteString(line + "\n"); err != nil {
						log.Printf("[-] 写入 %s/js.txt 失败: %v", folderName, err)
					}
				}

				if paramsRegex.MatchString(line) {
					fmt.Println(Yellow("[+] Find params：" + line))
					if _, err := paramsFile.WriteString(line + "\n"); err != nil {
						log.Printf("[-] 写入 %s/parameters.txt 失败: %v", folderName, err)
					}
				}

				if xlsRegex.MatchString(line) {
					fmt.Println(Cyan("[+] Find xls：" + line))
					if _, err := xlsFile.WriteString(line + "\n"); err != nil {
						log.Printf("[-] 写入 %s/xls.txt 失败: %v", folderName, err)
					}
				}

				if csvRegex.MatchString(line) {
					fmt.Println(Purple("[+] Find csv：" + line))
					if _, err := csvFile.WriteString(line + "\n"); err != nil {
						log.Printf("[-] 写入 %s/csv.txt 失败: %v", folderName, err)
					}
				}

				if pdfRegex.MatchString(line) {
					fmt.Println(Blue("[+] Find pdf：" + line))
					if _, err := pdfFile.WriteString(line + "\n"); err != nil {
						log.Printf("[-] 写入 %s/pdf.txt 失败: %v", folderName, err)
					}
				}
			}
		}(domain)
	}

	wg.Wait()
	fmt.Println(Green("[+] All write success."))
}
