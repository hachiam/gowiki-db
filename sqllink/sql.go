//@title sqllink
//@description 数据相关操作的包

package sqllink

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/denisenkom/go-mssqldb"
)

/*连接数据库的相关信息*/
var isdebug = true
var server = "localhost"
var port = 1433
var user = "sa"
var password = "19990901"
var database = "gowiki"

//数据的实体
type Paper struct {
	paperID int
	title   string
	species string
	body    string
}

func (p *Paper) GetPaperID() int {
	return p.paperID
}

func (p *Paper) GetTitle() string {
	return p.title
}

func (p *Paper) GetSpecies() string {
	return p.species
}

func (p *Paper) GetBody() string {
	return p.body
}

func (p *Paper) Save(title, body string) {
	conn := Connection()
	defer ConnectionClose(conn)
	UpdatePaper(conn, title, body)
}

func errhandler(description string, err error) {
	if err != nil {
		log.Fatal(description, err.Error())
		return
	}
}

func Connection() *sql.DB {
	//连接数据库
	var conString = fmt.Sprintf("server=%s;port%d;database=%s;user id=%s;password=%s", server, port, database, user, password)
	var conn, err = sql.Open("mssql", conString)
	errhandler("sql.Open", err)
	return conn
}

func ConnectionClose(conn *sql.DB) {
	//断开数据库连接
	conn.Close()
}

func InsertPaper(conn *sql.DB, title, body, species string) {
	//插入信息到数据库
	stmt, err := conn.Prepare("insert into paper values(?, ?, ?)")
	errhandler("InsertPaper : conn.Prepare", err)
	res, err := stmt.Exec(title, body, species)
	id, err := res.RowsAffected() //检查是否执行成功
	errhandler("InsertPaper : res.Rowsaffected", err)
	fmt.Println(id)
}

func SelectPaperbyTitle(conn *sql.DB, title string) *Paper {
	//通过title查询
	rows, err := conn.Query("select * from paper where title=?", title)
	errhandler("SelectPaperbyTitle : conn.Query", err)
	defer rows.Close()
	row := new(Paper)
	if rows.Next() {
		rows.Scan(&row.paperID, &row.title, &row.body, &row.species)
	}
	return row
}

func SelectAllPaper(conn *sql.DB) []*Paper {
	//查询所有的paper
	rows, err := conn.Query("select * from paper")
	errhandler("SelectAllPaper : conn.Query", err)
	var rowsData []*Paper
	for rows.Next() {
		row := new(Paper)
		rows.Scan(&row.paperID, &row.title, &row.body, &row.species)
		rowsData = append(rowsData, row)
	}
	return rowsData

}

func SelectBySpecies(conn *sql.DB, species string) []*Paper {
	//通过类别进行查询
	rows, err := conn.Query("select * from paper where species=?", species)
	errhandler("Selectbyspecies : conn.Query", err)
	var rowsData []*Paper
	for rows.Next() {
		row := new(Paper)
		rows.Scan(&row.paperID, &row.title, &row.body, &row.species)
		rowsData = append(rowsData, row)
	}
	return rowsData

}

func UpdatePaper(conn *sql.DB, title, body string) {
	//更新paper信息
	stmt, err := conn.Prepare("update paper set body=? where title=?")
	errhandler("UpdataPaper : conn.Prepare", err)
	res, err := stmt.Exec(body, title)
	errhandler("UpdatePaper : stmt.Exec", err)
	id, err := res.RowsAffected()
	errhandler("UpdataPaper : res.Rowsaffected", err)
	fmt.Println(id)
}

/*
func UpdateSpecies(conn *sql.DB, title, species string) {
	stmt, err := conn.Prepare("updata paper set species=? where title=?")
	errhandler("Updatespecies : conn.Prepare", err)
	res, err := stmt.Exec(species, title)
	errhandler("updatespecies : stmt.Exec", err)
	id, err := res.RowsAffected()
	errhandler("Updatespecies : res.Rowsaffected", err)
	fmt.Println(id)
}
*/

func DeletePaper(conn *sql.DB, title string) {
	//删除文章
	stmt, err := conn.Prepare("delete from paper where title=?")
	errhandler("Deletepaper : conn.Prepare", err)
	res, err := stmt.Exec(title)
	errhandler("Deletepaper : stmt.Exec", err)
	id, err := res.RowsAffected()
	errhandler("Deletepaper : res.Rowsaffected", err)
	fmt.Println(id)
}
