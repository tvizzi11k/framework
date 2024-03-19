package main

import (
 "database/sql"
 "fmt"
 "html/template"
 "log"
 "net/http"

 _ "github.com/go-sql-driver/mysql"
)

type Controller interface {
 HandleRequest(w http.ResponseWriter, req *http.Request)
}

type Model interface {
 Save(db Database) error
}

type Router struct {
 routes map[string]Controller
}

func (r *Router) HandleFunc(path string, controller Controller) {
 if r.routes == nil {
  r.routes = make(map[string]Controller)
 }
 r.routes[path] = controller
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
 path := req.URL.Path
 if controller, ok := r.routes[path]; ok {
  controller.HandleRequest(w, req)
 } else {
  http.NotFound(w, req)
 }
}

func NewRouter() *Router {
 return &Router{
  routes: make(map[string]Controller),
 }
}

type Database interface {
 Query(query string, args ...interface{}) (*sql.Rows, error)
 Exec(query string, args ...interface{}) (sql.Result, error)
}

func (u *User) Save(db Database) error {
 result, err := db.Exec(fmt.Sprintf("INSERT INTO %s (username, email) VALUES (?, ?)"))
 if err != nil {
  return err
 }
 lastInsertID, _ := result.LastInsertId()
 u.ID = int(lastInsertID)
 return nil
}

type GenericController struct {
 DB   Database
}

func (uc *UserController) HandleRequest(w http.ResponseWriter, req *http.Request) {
 table := req.FormValue("table")
 columns := req.FormValue("columns")
 values := req.FormValue("values")

 
 query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columns, values)

 result, err := gc.DB.Exec(query)
 if err != nil {
  http.Error(w, err.Error(), http.StatusInternalServerError)
  return
 }

 fmt.Fprintf(w, "Data inserted")
}

type View interface {
 Render(w http.ResponseWriter, data interface{}) error
}

type HTMLView struct {
 Template *template.Template
}

func (hv *HTMLView) Render(w http.ResponseWriter, data interface{}) error {
 err := hv.Template.Execute(w, data)
 if err != nil {
  return err
 }
 return nil
}

func NewHTMLView() *HTMLView {
 tmpl, err := template.ParseFiles("static/templates/index.html")
 if err != nil {
  log.Fatal("Error parsing HTML template:", err)
 }
 return &HTMLView{
  Template: tmpl,
 }
}

func main() {
 r := NewRouter()

 htmlView := NewHTMLView()

 db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/shedule")
 if err != nil {
  log.Fatal("Error connecting to database:", err)
 }

 genericController := &GenericController{
	 DB: db,
 }

 r.HandleFunc("/myAddress", genericController)

 server := &http.Server{
  Addr:    ":8080",
  Handler: r,
 }

 fmt.Println("Starting server on port 8080")
 err = server.ListenAndServe()
 if err != nil {
  log.Fatal("Server error", err)
 }
}

