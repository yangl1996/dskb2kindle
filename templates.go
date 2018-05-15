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

const mobiContents string = `<html>
  <head>
    <meta content="text/html; charset=utf-8" http-equiv="Content-Type"/>
    <title>Table of Contents</title>
  </head>
  <body>
    <h1>Contents</h1>
    {{#sections}}
    <h4>{{{title}}}</h4>
    <ul>
      {{#articles}}
      <li>
        <a href="{{file}}">{{{title}}}</a>
      </li>
      {{/articles}}
    </ul>
    {{/sections}}
  </body>
</html>`

const mobiNcx string = `<?xml version='1.0' encoding='utf-8'?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN" "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx xmlns:mbp="http://mobipocket.com/ns/mbp" xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1" xml:lang="en-US">
  <head>
    <meta content="{{doc_uuid}}" name="dtb:uid"/>
    <meta content="2" name="dtb:depth"/>
    <meta content="0" name="dtb:totalPageCount"/>
    <meta content="0" name="dtb:maxPageNumber"/>
  </head>
  <docTitle>
    <text>{{title}}</text>
  </docTitle>
  <docAuthor>
    <text>{{author}}</text>
  </docAuthor>
  <navMap>
    <navPoint playOrder="1" class="periodical" id="periodical">
      <mbp:meta-img src="{{masthead}}" name="mastheadImage"/>
      <navLabel><text>Table of Contents</text></navLabel>
      <content src="contents.html"/>
      {{#sections}}
      <navPoint playOrder="{{playorder}}" class="section" id="{{idref}}">
        <navLabel><text>{{title}}</text></navLabel>
        <content src="{{href}}"/>
        {{#articles}}
        <navPoint playOrder="{{playorder}}" class="article" id="{{idref}}">
          <navLabel><text>{{short_title}}</text></navLabel>
          <content src="{{href}}"/>
          <mbp:meta name="description">{{description}}</mbp:meta>
          <mbp:meta name="author">{{author}}</mbp:meta>
        </navPoint>
        {{/articles}}
      </navPoint>
      {{/sections}}
    </navPoint>
  </navMap>
</ncx>`

const mobiOpf string = `<?xml version='1.0' encoding='utf-8'?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="{{doc_uuid}}">
  <metadata>
    <meta content="cover-image" name="cover"/>
    <dc-metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
      <dc:title>{{title}}</dc:title>
      <dc:language>en-gb</dc:language>
      <dc:creator>{{author}}</dc:creator>
      <dc:publisher>{{publisher}}</dc:publisher>
      <dc:subject>{{subject}}</dc:subject>
      <dc:date>{{date}}</dc:date>
      <dc:description>{{description}}</dc:description>
    </dc-metadata>

    <x-metadata>
      <output content-type="application/x-mobipocket-subscription-magazine" encoding="utf-8"/>
    </x-metadata>


  </metadata>
  <manifest>
    <item href="contents.html" media-type="application/xhtml+xml" id="contents"/>
    <item href="nav-contents.ncx" media-type="application/x-dtbncx+xml" id="nav-contents"/>
    <item href="{{cover}}" media-type="{{cover_mimetype}}" id="cover-image"/>
    <item href="{{masthead}}" media-type="image/gif" id="masthead"/>
    {{#manifest_items}}
    <item href="{{href}}" media-type="{{media}}" id="{{idref}}"/>
    {{/manifest_items}}
  </manifest>
  <spine toc="nav-contents">
    <itemref idref="contents"/>
    {{#spine_items}}
    <itemref idref="{{idref}}"/>
    {{/spine_items}}
  </spine>
  <guide>
    <reference href="contents.html" type="toc" title="Table of Contents"/>
    {{#first_article}}
    <reference href="{{href}}" type="text" title="Beginning"/>
    {{/first_article}}
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
