package main

import (
    "fmt"
    "net/http"
    "log"
    "time"
    "golang.org/x/net/html"
    "strings"
    "strconv"
    "flag"
    "os"
    "io"
    "path/filepath"
    "text/template"
    "os/exec"
)

type Section struct {
    Title string
    Articles []string
    Path string
}

type Article struct {
    H1 string
    H2 string
    H3 string
    H4 string
    Text []Token
    Path string
}

type Token struct {
    Para string
    Image string
}

type FileEntry struct {
    Path string
    Title string
    Playorder string
    Idref string
}

type Manifest struct {
    Sections []ManifestSection
    Images []FileEntry
}

type ManifestSection struct {
    Self FileEntry
    Articles []FileEntry
}

type BookMetadata struct {
    Uuid string
    Title string
    Author string
    Masthead string
    Manifest Manifest
    Date string
    Cover string
}

func main() {
    /* parse command line arguments */
    dateArg := flag.String("date", "today", "the date of the issue to be fetched, in format YYYY-MM-DD, MM-DD, or DD")
    workspaceArg := flag.String("workspace", "./dskb2kindle", "directory to store temporary files and results")
    outputArg := flag.String("output", "dushikuaibao.pobi", "output filename")
    flag.Parse()
    log.Println("Process started")
    _, err := exec.LookPath("kindlegen")
    if err != nil {
        log.Fatalln("Kindlegen is missing")
    }
    workspacePath, err := filepath.Abs(*workspaceArg)
    if err != nil {
        log.Fatalln("Error parsing workspace path")
    }
    if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
        err = os.Mkdir(workspacePath, os.ModePerm)
        if err != nil {
            log.Fatalf("Error creating %s\n", workspacePath)
        }
    } else {
        log.Fatalf("%s already exists\n", workspacePath)
    }
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
    dskbMastheadURL := "http://hzdaily.hangzhou.com.cn/img/logo/dskb2.png"

    /* get and parse table of content */
    actionFunc, textActFunc, tableOfContentResultsRetriever := tableOfContentParser(dskbBaseURL)
    parseURL(dskbURL, actionFunc, textActFunc)
    tableOfContent := tableOfContentResultsRetriever()

    /* get and parse the thumbnail of the frontpage */
    actionFunc, textActFunc, frontPageResultsRetriever := frontPageParser()
    parseURL(dskbFrontPageURL, actionFunc, textActFunc)
    frontPageImageURL := frontPageResultsRetriever()

    /* download thumbnail of the frontpage */
    thumbnailPath := filepath.Join(workspacePath, "thumbnail.jpg")
    resp, err := http.Get(frontPageImageURL)
    if err != nil {
        log.Fatalln("Error downloading thumbnail")
    }
    thumbnailFile, err := os.Create(thumbnailPath)
    if err != nil {
        log.Fatalln("Error creating thumbnail file")
    }
    _, err = io.Copy(thumbnailFile, resp.Body)
    if err != nil {
        log.Fatalln("Error writing to thumbnail file")
    }
    thumbnailFile.Close()
    resp.Body.Close()

    /* download masthead */
    mastheadPath := filepath.Join(workspacePath, "masthead.png")
    resp, err = http.Get(dskbMastheadURL)
    if err != nil {
        log.Fatalln("Error downloading masthead")
    }
    mastheadFile, err := os.Create(mastheadPath)
    if err != nil {
        log.Fatalln("Error creating masthead file")
    }
    _, err = io.Copy(mastheadFile, resp.Body)
    if err != nil {
        log.Fatalln("Error writing to masthead file")
    }
    mastheadFile.Close()
    resp.Body.Close()

    /* create templates */
    articleTemplate := template.Must(template.New("article").Parse(mobiArticle))
    sectionTemplate := template.Must(template.New("section").Parse(mobiSection))
    contentsTemplate := template.Must(template.New("contents").Parse(mobiContents))
    ncxTemplate := template.Must(template.New("ncx").Parse(mobiNcx))
    opfTemplate := template.Must(template.New("opf").Parse(mobiOpf))

    /* prepare manifest list */
    var manifest Manifest
    getNextPlayOrder := getIncreasingInt(2)
    getNextIdRef := getIncreasingInt(1)

    /* get and parse each article, first go through each section */
    for sectionIdx, section := range tableOfContent {
        tableOfContent[sectionIdx].Path = filepath.Join(workspacePath, strconv.Itoa(sectionIdx))
        section.Path = filepath.Join(workspacePath, strconv.Itoa(sectionIdx))
        err = os.Mkdir(section.Path, os.ModePerm)
        if err != nil {
            log.Fatalf("Error creating directory %s\n", section.Path)
        }
        sectionFile, err := os.Create(filepath.Join(section.Path, "section.html"))
        if err != nil {
            log.Fatalln("Error creating section HTML file")
        }
        err = sectionTemplate.Execute(sectionFile, section)
        sectionFile.Close()
        if err != nil {
            log.Fatalln("Error applying template to section")
        }
        manifest.Sections = append(manifest.Sections, ManifestSection{FileEntry{filepath.Join(section.Path, "section.html"), section.Title, getNextPlayOrder(), getNextIdRef()}, []FileEntry{}})
        for articleIdx, articleURL := range section.Articles {
            elementAct, textAct, articleResultsRetriever := articleParser()
            parseURL(articleURL, elementAct, textAct)
            article := articleResultsRetriever()
            article.Path = filepath.Join(section.Path, strings.Join([]string{strconv.Itoa(articleIdx), ".html"}, ""))
            for tokenIdx, token := range article.Text {
                if token.Image != "" {
                    imagePath := filepath.Join(section.Path, strings.Join([]string{"img", strconv.Itoa(articleIdx), strconv.Itoa(tokenIdx), ".jpg"}, "_"))
                    resp, err := http.Get(token.Image)
                    if err != nil {
                        log.Fatalln("Error downloading image")
                    }
                    imageFile, err := os.Create(imagePath)
                    if err != nil {
                        log.Fatalln("Error creating image file")
                    }
                    _, err = io.Copy(imageFile, resp.Body)
                    if err != nil {
                        log.Fatalln("Error writing to image file")
                    }
                    imageFile.Close()
                    resp.Body.Close()
                    article.Text[tokenIdx].Image = imagePath
                    token.Image = imagePath
                    manifest.Images = append(manifest.Images, FileEntry{imagePath, "", "", getNextIdRef()})
                }
            }
            htmlFile, err := os.Create(article.Path)
            if err != nil {
                log.Fatalln("Error creating article HTML file")
            }
            err = articleTemplate.Execute(htmlFile, article)
            htmlFile.Close()
            if err != nil {
                log.Fatalln("Error applying template to article")
            }
            manifest.Sections[len(manifest.Sections)-1].Articles = append(manifest.Sections[len(manifest.Sections)-1].Articles, FileEntry{article.Path, article.H1, getNextPlayOrder(), getNextIdRef()})
        }
    }

    contentsFile, err := os.Create(filepath.Join(workspacePath, "contents.html"))
    if err != nil {
        log.Fatalln("Error creating contents file")
    }
    err = contentsTemplate.Execute(contentsFile, manifest)
    contentsFile.Close()
    if err != nil {
        log.Fatalln("Error applying template to table of contents")
    }

    /* prepare metadata */
    var metadata BookMetadata
    metadata.Manifest = manifest
    metadata.Uuid = strings.Join([]string{"dushikuaibao.12345", hztime.Format("2006-01-02")}, "-")
    metadata.Title = strings.Join([]string{"都市快报", hztime.Format("2006-01-02")}, " ")
    metadata.Author = "杭州日报报业集团"
    metadata.Masthead = mastheadPath
    metadata.Date = hztime.Format("2006-01-02")
    metadata.Cover = thumbnailPath

    ncxFile, err := os.Create(filepath.Join(workspacePath, "nav-contents.ncx"))
    if err != nil {
        log.Fatalln("Error creating NCX file")
    }
    err = ncxTemplate.Execute(ncxFile, metadata)
    ncxFile.Close()
    if err != nil {
        log.Fatalln("Error applying template to NCX")
    }

    opfFile, err := os.Create(filepath.Join(workspacePath, "dskb2kindle.opf"))
    if err != nil {
        log.Fatalln("Error creating OPF file")
    }
    err = opfTemplate.Execute(opfFile, metadata)
    opfFile.Close()
    if err != nil {
        log.Fatalln("Error applying template to OPF")
    }

    /* call kindlegen */
    cmd := exec.Command("kindlegen", "-o", *outputArg, filepath.Join(workspacePath, "dskb2kindle.opf"))
    err = cmd.Run()
    if err != nil {
        log.Fatalln("Kindlegen returned with error")
    } else {
        log.Printf("Successfully generated at %s\n", filepath.Join(workspacePath, *outputArg))
    }
}

func parseURL(url string, act func(*html.Node), textAct func(*html.Node)) {
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
    var processHTML func (*html.Node, func(*html.Node), func(*html.Node))
    processHTML = func (n *html.Node, act func(*html.Node), textAct func(*html.Node)) {
        switch n.Type {
        case html.ErrorNode:
            log.Fatalln("Error parsing DOM node")
        case html.ElementNode:
            act(n)
        case html.TextNode:
            textAct(n)
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            processHTML(c, act, textAct)
        }
    }
    processHTML(doc, act, textAct)
}

func tableOfContentParser(baseURL string) (func(*html.Node), func(*html.Node), func() []Section) {
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
                tableOfContent = append(tableOfContent, Section{strings.Trim(n.FirstChild.Data, " "), []string{}, ""})
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
    processText := func(n *html.Node) {
        return
    }
    getResults := func() []Section {
        log.Printf("Found %d sections", len(tableOfContent))
        return tableOfContent
    }
    return processTree, processText, getResults
}

func frontPageParser() (func(*html.Node), func(*html.Node), func() string) {
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
    processText := func(n *html.Node) {
        return
    }
    getResults := func() string {
        if parsingState != 2 {
            log.Fatalln("Error parsing front page thumbnail URL")
        }
        return frontPageImageURL
    }
    return processFrontPage, processText, getResults
}

func articleParser() (func(*html.Node), func(*html.Node), func() Article) {
    var article Article
    parsingState := 0
    processElement := func (n *html.Node) {
        switch n.Data {
        case "div":
            for _, a := range n.Attr {
                switch {
                case a.Key == "class" && a.Val == "content":
                    parsingState = 3
                case a.Key == "class" && a.Val == "head":
                    parsingState = 1
                }
            }
        case "img":
            if parsingState == 3 {
                for _, a := range n.Attr {
                    switch {
                    case a.Key == "src":
                        article.Text = append(article.Text, Token{"", a.Val})
                    }
                }
            }
        case "p":
            if parsingState == 1 {
                parsingState = 2
            } else if parsingState == 2 {
                parsingState = 0
            }
        }
    }
    processText := func (n *html.Node) {
        if parsingState == 1 || parsingState == 2 {
            switch  n.Parent.Data {
            case "h1":
                article.H1 = strings.Trim(n.Data, " \n\t　")
            case "h2":
                article.H2 = strings.Trim(n.Data, " \n\t　")
            case "h3":
                article.H3 = strings.Trim(n.Data, " \n\t　")
            case "p":
                article.H4 = strings.Trim(n.Data, " \n\t　")
                parsingState = 0
            }
        } else if parsingState == 3 {
            if n.Parent.Data == "p" {
                stripped := strings.Trim(n.Data, " \n\t　")
                if stripped != "" {
                    article.Text = append(article.Text, Token{stripped, ""})
                }
            }
        }
    }
    getResults := func() Article {
        return article
    }
    return processElement, processText, getResults
}

func getIncreasingInt(start int) func() string {
    current := start - 1
    getNext := func () string {
        current += 1
        nextString := strconv.Itoa(current)
        return nextString
    }
    return getNext
}

