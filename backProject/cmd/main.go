package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"html/template"
	"log"
	"net/http"
)

type Controller interface {
	HandleRequest(w http.ResponseWriter, req *http.Request)
}

type Model interface {
	Save() error
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

func (u *User) Save() error {
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/shedule") // сюда вносим данные о бд
	if err != nil {
		return err
	}
	defer func() {
		if dErr := db.Close(); dErr != nil {
			log.Println("Error closing db:", dErr)
		}
	}()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	_, err = db.Exec("INSERT INTO users (username, email) VALUES (?, ?)", u.Username, u.Email)
	if err != nil {
		return err
	}
	return nil
}

type UserController struct {
	Model Model
	View  View
}

func (uc *UserController) HandleRequest(w http.ResponseWriter, req *http.Request) {
	username := req.FormValue("username")
	email := req.FormValue("email")

	newUser := &User{
		Username: username,
		Email:    email,
	}
	if err := newUser.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := uc.View.Render(w, nil)
	if err != nil {
		http.Error(w, "Error rend html temp", http.StatusInternalServerError)
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

type AddUserController struct {
	Model Model
}

func (ac *AddUserController) HandleRequest(w http.ResponseWriter, req *http.Request) {
	username := req.FormValue("username")
	email := req.FormValue("email")

	newUser := &User{
		Username: username,
		Email:    email,
	}
	if err := newUser.Save(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, req, "/myAddress", http.StatusFound)
}

func main() {
	r := NewRouter()

	htmlView := NewHTMLView()

	userController := &UserController{
		Model: &User{},
		View:  htmlView,
	}

	r.HandleFunc("/myAddress", userController)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	addUserController := &AddUserController{
		Model: &User{},
	}
	r.HandleFunc("/addUser", addUserController)

	fmt.Println("Starting server on port 8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Server error", err)
	}
}
