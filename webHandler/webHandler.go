package webHandler

import (
	"bytes"
	"errors"
	"fmt"
	"gowiki-db/sqllink"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

var templates = template.Must(template.ParseFiles("./tmpl/edit.html", "./tmpl/view.html", "./tmpl/addFile.html", "./tmpl/index.html", "./tmpl/list.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|add|list|download|delete)/([a-zA-Z0-9\\x{4e00}-\\x{9fa5}]+)$")

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("invalid Page Title")
	}
	return m[2], nil //the title is the second subexpression.
}

func loadPage(title string) *sqllink.Paper {
	conn := sqllink.Connection()
	defer sqllink.ConnectionClose(conn)
	p := sqllink.SelectPaperbyTitle(conn, title)
	return p
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *sqllink.Paper) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	//主页
	renderTemplate(w, "index", nil)
	// t, err := template.ParseFiles("./tmpl/index.html")
	// if err != nil {
	// 	log.Fatal("index : template.PatseFiles", err.Error())
	// }
	// err = t.Execute(w, nil)
	// if err != nil {
	// 	log.Fatal("index : t.Execute", err.Error())
	// }
}

func View(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		log.Fatal("View : gettitle", err.Error())
	}
	p := loadPage(title)
	renderTemplate(w, "view", p)
	// t, err := template.ParseFiles("./tmpl/view.html")
	// if err != nil {
	// 	log.Fatal("View : template.Parsefiles", err.Error())
	// 	return
	// }
	// err = t.Execute(w, p)
	// if err != nil {
	// 	log.Fatal("View : t.Execute", err.Error())
	// }
}

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

func AddPaper(w http.ResponseWriter, r *http.Request) {
	//实现文件的增添
	if r.Method == "GET" {
		renderTemplate(w, "addFile", nil)
		// t, err := template.ParseFiles("./tmpl/addFile.html")
		// if err != nil {
		// 	log.Fatal(err)
		// 	return
		// }
		// err = t.Execute(w, nil)
		// if err != nil {
		// 	log.Fatal(err)
		// 	return
		// }
	} else if r.Method == "POST" {
		r.ParseForm()
		title := r.Form.Get("title")
		species := r.Form.Get("species")
		conn := sqllink.Connection()
		defer sqllink.ConnectionClose(conn)
		p := sqllink.SelectPaperbyTitle(conn, title)
		if p.GetTitle() == ""{
			sqllink.InsertPaper(conn, title, "新文件", species)
		} 
		http.Redirect(w, r, "/edit/"+title, 302)
	}

}

func List(w http.ResponseWriter, r *http.Request) {
	html := loadHTML("./tmpl/list.html")
	conn := sqllink.Connection()
	defer sqllink.ConnectionClose(conn)
	papers := sqllink.SelectAllPaper(conn)
	temp := ""
	if papers[0] != nil {
		for i := 0; i < len(papers); i++ {
			temp += `<li><a href="/view/` + papers[i].GetTitle() + `">` + papers[i].GetTitle() + `</a></li>`

		}
	}
	html = bytes.Replace(html, []byte("@html"), []byte(temp), 1)
	w.Write(html)

}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		log.Fatal("DownloadFile : getTitle", err.Error())
	}
	p := loadPage(title)
	f, err := os.Create(title + ".md")
	defer f.Close()
	if err != nil {
		log.Fatal("DownloadFile : os.Create", err.Error())
	} else {
		_,err = f.Write([]byte(p.GetBody()))
		if err != nil {
			log.Fatal("Downloadfile : f.Write ", err.Error()) 
		}
	}
}

func DeletePaper(w http.ResponseWriter, r *http.Request) {
	title, err := getTitle(w, r)
	if err != nil {
		log.Fatal("Deletepaper : gettitle", err.Error())
	}
	conn := sqllink.Connection()
	defer sqllink.ConnectionClose(conn)
	sqllink.DeletePaper(conn, title)
	w.Write([]byte("<html><h1><a href="+`/list`+">点击文件列表</a></h1></html>"))
}

///-------------------测试
func Test() {
	conn := sqllink.Connection()
	defer sqllink.ConnectionClose(conn)
	papers := sqllink.SelectAllPaper(conn)
	fmt.Println(papers[0])
	/*	for i := 0; i < len(papers); i++{
			fmt.Printf("title: %s, \tbody: %s, \t", papers[i].GetTitle(), papers[i].GetBody())
		}
	*/
}
