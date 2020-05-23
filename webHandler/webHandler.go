package webHandler

import (
	"bytes"
	"errors"
	"fmt"
	"gowiki-db/sqllink"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

//预先加载全部模板
var templates = template.Must(template.ParseFiles("./tmpl/edit.html", "./tmpl/view.html", "./tmpl/addFile.html", "./tmpl/index.html", "./tmpl/list.html", "./tmpl/addMarkdown.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|add|list|download|delete|addmarkdown)/([a-zA-Z0-9\\x{4e00}-\\x{9fa5}]+)$")

//@title getTitle
//@description 通过正则表达式从URL中得到文章标题
//@return void
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil //the title is the second subexpression.
}

//@title loadPage
//@description 通过文章标题得到指定的Paper对象
//@return sqllink.Paper，即保存文章信息的实体
func loadPage(title string) *sqllink.Paper {
	conn := sqllink.Connection()
	defer sqllink.ConnectionClose(conn)
	p := sqllink.SelectPaperbyTitle(conn, title)
	return p
}

//@title renderTemplate
//@description 从已加载模板中读取指定的模板
//@return void
func renderTemplate(w http.ResponseWriter, tmpl string, p *sqllink.Paper) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//@title Index
//@description 首页
//@return void
func Index(w http.ResponseWriter, r *http.Request) {
	//主页
	renderTemplate(w, "index", nil)
}

//@title View
//@description 对单片文章的展示
//@return void
func View(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		log.Fatal("View : gettitle", err.Error())
	}
	p := loadPage(title)
	renderTemplate(w, "view", p)
}

//@title EditHandler
//@description 编辑文章
//@return void
func EditHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	p := loadPage(title)
	if p.GetTitle() == "" {
		conn := sqllink.Connection()
		defer sqllink.ConnectionClose(conn)
		sqllink.InsertPaper(conn, title, "新文件", "新类别")
	}
	if err != nil {
		return
	}
	renderTemplate(w, "edit", p)
}

//@title loadHTML
//@description 从模板库中读取指定的html模版
//@return 返回保存文件信息的字节流
func loadHTML(name string) []byte {
	f, err := os.Open(name)
	if err != nil {
		return []byte("<html><head></head><body><h3>Errors</h3></body></html>")
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return []byte("<html><head></head><body><h3>Errors</h3></body></html>")
	}
	return buf
}

//@title SaveHandler
//@description 保存文章
//@return void
func SaveHandler(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		return
	}
	body := r.FormValue("body")
	p := new(sqllink.Paper)
	p.Save(title, body)
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

//@title AddPaper
//@description 添加文章
//@return void
func AddPaper(w http.ResponseWriter, r *http.Request) {
	//实现文件的增添
	if r.Method == "GET" {
		renderTemplate(w, "addFile", nil)
	} else if r.Method == "POST" {
		r.ParseForm()
		title := r.Form.Get("title")
		species := r.Form.Get("species")
		conn := sqllink.Connection()
		defer sqllink.ConnectionClose(conn)
		p := sqllink.SelectPaperbyTitle(conn, title)
		if p.GetTitle() == "" {
			sqllink.InsertPaper(conn, title, "新文件", species)
		}
		http.Redirect(w, r, "/edit/"+title, 302)
	}

}

//@title List
//@description 从数据库读取文章，并在页面展示
//@return void
func List(w http.ResponseWriter, r *http.Request) {
	html := loadHTML("./tmpl/list.html")
	conn := sqllink.Connection()
	defer sqllink.ConnectionClose(conn)
	papers := sqllink.SelectAllPaper(conn)
	temp := ""
	if papers == nil {
		w.Write([]byte("<html><h1>文章列表为空，请先添加文章！！</h1><hr/><br/><a href=/add>添加文章</a></html>"))
		return
	}
	if papers[0] != nil {
		for i := 0; i < len(papers); i++ {
			temp += `<li><a href="/view/` + papers[i].GetTitle() + `">` + papers[i].GetTitle() + `</a></li>`

		}
	}
	html = bytes.Replace(html, []byte("@html"), []byte(temp), 1)
	w.Write(html)

}

//@title DownloadFile
//@description 下载文章到本地
//@return void
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		log.Fatal("DownloadFile : getTitle", err.Error())
	}
	p := loadPage(title)
	f, err := os.Create("./data/" + title + ".md")
	defer f.Close()
	if err != nil {
		log.Fatal("DownloadFile : os.Create", err.Error())
	} else {
		_, err = f.Write([]byte(p.GetBody()))
		if err != nil {
			log.Fatal("Downloadfile : f.Write ", err.Error())
		}
	}
	w.Write([]byte("<html><h2>导出文件成功</h2><a href=/list>返回文件列表</a></html>"))
}

//@title DeletePaper
//@description 从数据库删除文章
//@return void
func DeletePaper(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		log.Fatal("Deletepaper : gettitle", err.Error())
	}
	conn := sqllink.Connection()
	defer sqllink.ConnectionClose(conn)
	sqllink.DeletePaper(conn, title)
	w.Write([]byte("<html><h1><a href=" + `/list` + ">点击文件列表</a></h1></html>"))
}

//@title UploadMarkdown
//@description 控制上传markdown的行为
//@return void
func UploadMarkdown(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "addMarkdown", nil)
		return
	} else if r.Method == "POST" {

		f, h, err := r.FormFile("file")
		if err != nil {
			w.Write([]byte("文件上传有误 ：" + err.Error()))
			return
		}
		t := h.Header.Get("Content-Type")
		log.Println(t)
		fileExt := path.Ext(h.Filename)
		if strings.EqualFold(fileExt, ".md") || strings.EqualFold(fileExt, ".markdown") {
			fmt.Println("addMarkdownfile...")
		} else {
			w.Write([]byte("<html><h3>上传文件必须是markdown文件</h3><a href=/addmarkdown>返回</a></html>"))
			return
		}
		out, err := os.Create("./data/" + h.Filename)
		if err != nil {
			io.WriteString(w, "文件创建失败:"+err.Error())
			return
		}
		_, err = io.Copy(out, f)
		if err != nil {
			io.WriteString(w, "文件保存失败:"+err.Error())
			return
		}
		fileNameOnly := strings.TrimSuffix(path.Base(h.Filename), path.Ext(h.Filename))
		buf, err := ioutil.ReadFile("./data/" + h.Filename)
		if err != nil {
			w.Write([]byte("<html><h1>文件上传有误</h1></html>"))
		}
		//插入数据库
		conn := sqllink.Connection()
		defer sqllink.ConnectionClose(conn)
		p := sqllink.SelectPaperbyTitle(conn, fileNameOnly)
		if p.GetTitle() == "" {
			sqllink.InsertPaper(conn, fileNameOnly, string(buf), "新类别")
		}
		http.Redirect(w, r, "/view/"+fileNameOnly, http.StatusFound)
		return
	}
}
