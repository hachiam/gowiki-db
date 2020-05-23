package main

import (
	"gowiki-db/webHandler"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", webHandler.Index)
	http.HandleFunc("/view/", webHandler.View)
	http.HandleFunc("/list", webHandler.List)
	http.HandleFunc("/edit/", webHandler.EditHandler)
	http.HandleFunc("/save/", webHandler.SaveHandler)
	http.HandleFunc("/add", webHandler.AddPaper)
	http.HandleFunc("/download/", webHandler.DownloadFile)
	http.HandleFunc("/delete/", webHandler.DeletePaper)
	http.HandleFunc("/addmarkdown", webHandler.UploadMarkdown)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(":9090", nil))
}
