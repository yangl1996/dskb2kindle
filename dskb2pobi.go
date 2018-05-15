package main

import (
    "fmt"
    "net/http"
    "log"
    "time"
    "golang.org/x/net/html"
    "strings"
    "flag"
)

type Section struct {
    Title string
    Articles []string
}

func main() {
    /* parse command line arguments */
    dateArg := flag.String("date", "today", "the date of the issue to be fetched, in format YYYY-MM-DD, MM-DD, or DD")
    flag.Parse()
    log.Println("Process started")
    hangzhou := time.FixedZone("Hangzhou Time", int((8 * time.Hour).Seconds()))
    var hztime time.Time
    if *dateArg == "today" {
        hztime = time.Now().In(hangzhou)
    } else {
        t1, err1 := time.ParseInLocation("2006-01-02", *dateArg, hangzhou)
        t2, err2 := time.ParseInLocation("01-02", *dateArg, hangzhou)
        t3, err3 := time.ParseInLocation("02", *dateArg, hangzhou)
        switch {
        case err1 == nil:
            hztime = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, hangzhou)
        case err2 == nil:
            hztime = time.Date(time.Now().In(hangzhou).Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, hangzhou)
        case err3 == nil:
            hztime = time.Date(time.Now().In(hangzhou).Year(), time.Now().In(hangzhou).Month(), t3.Day(), 0, 0, 0, 0, hangzhou)
        default:
            log.Fatalln("Error parsing date option")
        }
    }
    log.Printf("Retrieving Dushikuaibao issue for %s\n", hztime.Format("Jan 2 2006"))

    /* format URLs */
    dskbURL := fmt.Sprintf("http://mdaily.hangzhou.com.cn/dskb/%s/article_list_%s.html", hztime.Format("2006/01/02"), hztime.Format("20060102"))
    dskbFrontPageURL := fmt.Sprintf("http://mdaily.hangzhou.com.cn/dskb/%s/page_list_%s.html", hztime.Format("2006/01/02"), hztime.Format("20060102"))
    dskbBaseURL := fmt.Sprintf("http://mdaily.hangzhou.com.cn/dskb/%s/", hztime.Format("2006/01/02"))

    /* get and parse table of content */
    actionFunc, tableOfContentResultsRetriever := tableOfContentParser(dskbBaseURL)
    parseURL(dskbURL, actionFunc)
    tableOfContent := tableOfContentResultsRetriever()

    /* get and parse the thumbnail of the frontpage */
    actionFunc, frontPageResultsRetriever := frontPageParser() 
    parseURL(dskbFrontPageURL, actionFunc)
    frontPageImageURL := frontPageResultsRetriever()

    fmt.Println(frontPageImageURL)
    fmt.Println(tableOfContent)
}

func parseURL(url string, act func(*html.Node)) {
    resp, err := http.Get(url)
    if err != nil {
        log.Fatalln("Error communicating with Dushikuaibao server")
    }
    doc, err := html.Parse(resp.Body)
    if err != nil {
        resp.Body.Close()
        log.Fatalln("Error parsing HTML")
    }
    resp.Body.Close()
    var processHTML func (*html.Node, func(*html.Node))
    processHTML = func (n *html.Node, act func(*html.Node)) {
        switch n.Type {
        case html.ErrorNode:
            log.Fatalln("Error parsing DOM node")
        case html.ElementNode:
            act(n)
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            processHTML(c, act)
        }
    }
    processHTML(doc, act)
}

func tableOfContentParser(baseURL string) (func(*html.Node), func() []Section) {
    var tableOfContent []Section
    parsingState := 0
    processTree := func(n *html.Node) {
        switch n.Data {
        case "title":
            if n.FirstChild.Data == "404页面" {
                log.Fatalln("HTTP 404, this issue is not available")
            }
        case "div":
            if len(n.Attr) == 0 {
                break
            }
            if n.Attr[0].Key == "class" && n.Attr[0].Val == "title" {
                if n.FirstChild.Data  == " 第A01版：都市快报" {
                    break
                }
                tableOfContent = append(tableOfContent, Section{strings.Trim(n.FirstChild.Data, " "), []string{}})
                parsingState = 1
            }
        case "a":
            if parsingState == 0 {
                break
            }
            if len(n.Attr) == 0 {
                break
            }
            tableOfContent[len(tableOfContent)-1].Articles = append(tableOfContent[len(tableOfContent)-1].Articles, strings.Join([]string{baseURL, n.Attr[0].Val}, ""))
        }
    }
    getResults := func() []Section {
        log.Printf("Found %d sections", len(tableOfContent))
        return tableOfContent
    }
    return processTree, getResults
}

func frontPageParser() (func(*html.Node), func() string) {
    var frontPageImageURL string
    parsingState := 0
    processFrontPage := func(n *html.Node) {
        switch n.Data {
        case "title":
            if n.FirstChild.Data == "404页面" {
                log.Fatalln("HTTP 404, error retrieving front page")
            }
        case "div":
            for _, a := range n.Attr {
                if a.Key == "class" && a.Val == "section page1" {
                    parsingState = 1
                }
            }
        case "img":
            if parsingState == 1 {
                for _, a := range n.Attr {
                    if a.Key == "data-src" {
                        frontPageImageURL = a.Val
                        parsingState = 2
                    }
                }
            }
        }
    }
    getResults := func() string {
        if parsingState != 2 {
            log.Fatalln("Error parsing front page thumbnail URL")
        }
        return frontPageImageURL
    }
    return processFrontPage, getResults
}
