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

type User struct {
	ID       int
	Username string
	Email    string
}

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (u *User) Save(db Database) error {
	result, err := db.Exec("INSERT INTO users (username, email) VALUES (?, ?)", u.Username, u.Email)
	if err != nil {
		return err
	}
	lastInsertID, _ := result.LastInsertId()
	u.ID = int(lastInsertID)
	return nil
}

type UserController struct {
	DB   Database
	View View
}

func (uc *UserController) HandleRequest(w http.ResponseWriter, req *http.Request) {
	username := req.FormValue("username")
	email := req.FormValue("email")

	newUser := &User{
		Username: username,
		Email:    email,
	}
	if err := newUser.Save(uc.DB); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := uc.View.Render(w, newUser)
	if err != nil {
		http.Error(w, "Error rendering HTML template", http.StatusInternalServerError)
	}
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

	userController := &UserController{
		DB:   db,
		View: htmlView,
	}

	r.HandleFunc("/myAddress", userController)

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
