package main

const mobiArticle string = `<!DOCTYPE html>
<html lang="zh">
  <head>
    <meta charset="utf-8">
    <title>{{.H1}}</title>
  </head>
  <body>
  <h1>{{.H1}}</h1>
  <h2>{{.H2}}</h2>
  <h3>{{.H3}}</h3>
  <h4>{{.H4}}</h4>
  {{range .Text}}
  {{if ne .Para ""}}
  <p>{{.Para}}</p>
  {{else if ne .Image ""}}
  <img src="{{.Image}}">
  {{end}}
  {{end}}
  </body>
</html>
`

const mobiContents string = `<!DOCTYPE html>
<html lang="zh">
  <head>
    <meta content="text/html; charset=utf-8" http-equiv="Content-Type"/>
    <title>目录</title>
  </head>
  <body>
    <h1>本期内容</h1>
    {{range .Sections}}
    <h4>{{.Self.Title}}</h4>
    <ul>
      {{range .Articles}}
      <li>
        <a href="{{.Path}}">{{.Title}}</a>
      </li>
      {{end}}
    </ul>
    {{end}}
  </body>
</html>`

const mobiNcx string = `<?xml version='1.0' encoding='utf-8'?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx xmlns:mbp="http://mobipocket.com/ns/mbp" xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1" xml:lang="zh-CN">
  <head>
    <meta content="{{.Uuid}}" name="dtb:uid"/>
    <meta content="2" name="dtb:depth"/>
    <meta content="0" name="dtb:totalPageCount"/>
    <meta content="0" name="dtb:maxPageNumber"/>
  </head>
  <docTitle>
    <text>{{.Title}}</text>
  </docTitle>
  <docAuthor>
    <text>{{.Author}}</text>
  </docAuthor>
  <navMap>
    <navPoint playOrder="1" class="periodical" id="periodical">
      <mbp:meta-img src="{{.Masthead}}" name="mastheadImage"/>
      <navLabel><text>目录</text></navLabel>
      <content src="contents.html"/>
      {{range .Manifest.Sections}}
      <navPoint playOrder="{{.Self.Playorder}}" class="section" id="{{.Self.Idref}}">
        <navLabel><text>{{.Self.Title}}</text></navLabel>
        <content src="{{.Self.Path}}"/>
        {{range .Articles}}
        <navPoint playOrder="{{.Playorder}}" class="article" id="{{.Idref}}">
          <navLabel><text>{{.Title}}</text></navLabel>
          <content src="{{.Path}}"/>
        </navPoint>
        {{end}}
      </navPoint>
      {{end}}
    </navPoint>
  </navMap>
</ncx>`

const mobiOpf string = `<?xml version='1.0' encoding='utf-8'?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="{{.Uuid}}">
  <metadata>
    <meta content="cover-image" name="cover"/>
    <dc-metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
      <dc:title>{{.Title}}</dc:title>
      <dc:language>en-gb</dc:language>
      <dc:creator>{{.Author}}</dc:creator>
      <dc:publisher>dskb2kindle (Unofficial)</dc:publisher>
      <dc:subject>News</dc:subject>
      <dc:date>{{.Date}}</dc:date>
      <dc:description>Dushikuaibao unofficially generated from the web version</dc:description>
    </dc-metadata>

    <x-metadata>
      <output content-type="application/x-mobipocket-subscription-magazine" encoding="utf-8"/>
    </x-metadata>


  </metadata>
  <manifest>
    <item href="contents.html" media-type="application/xhtml+xml" id="contents"/>
    <item href="nav-contents.ncx" media-type="application/x-dtbncx+xml" id="nav-contents"/>
    <item href="{{.Cover}}" media-type="image/jpg" id="cover-image"/>
    <item href="{{.Masthead}}" media-type="image/png" id="masthead"/>
    {{range .Manifest.Images}}
    <item href="{{.Path}}" media-type="image/jpg" id="{{.Idref}}"/>
    {{end}}
    {{range .Manifest.Sections}}
    <item href="{{.Self.Path}}" media-type="application/xhtml+xml" id="{{.Self.Idref}}"/>
    {{range .Articles}}
    <item href="{{.Path}}" media-type="application/xhtml+xml" id="{{.Idref}}"/>
    {{end}}
    {{end}}
  </manifest>
  <spine toc="nav-contents">
    <itemref idref="contents"/>
    {{range .Manifest.Sections}}
    <itemref idref="{{.Self.Idref}}"/>
    {{range .Articles}}
    <itemref idref="{{.Idref}}"/>
    {{end}}
    {{end}}
  </spine>
  <guide>
    <reference href="contents.html" type="toc" title="Table of Contents"/>
    <reference href="{{(index (index .Manifest.Sections 0).Articles 0).Path}}" type="text" title="Beginning"/>
  </guide>
</package>`

var mobiSection string = `<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xml:lang="zh">
  <head>
    <meta content="http://www.w3.org/1999/xhtml; charset=utf-8" http-equiv="Content-Type"/>
    <title>{{.Title}}</title>
  </head>
  <body>
    <p>&nbsp;</p>
    <p>&nbsp;</p>
    <p>&nbsp;</p>
    <p>&nbsp;</p>
    <h1>{{.Title}}</h1>
  </body>
</html>`
