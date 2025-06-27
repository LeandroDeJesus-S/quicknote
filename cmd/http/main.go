package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/LeandroDeJesus-S/quicknote/config"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles(
		"view/templates/base.html",
		"view/templates/pages/home.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(tpl.DefinedTemplates())

	err = tpl.ExecuteTemplate(w, "base.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func listNotes(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path, r.URL.RawQuery)

	w.Header().Set("teste", "123")
	w.Header().Set("teste", "456")
	fmt.Fprint(w, "List Notes")
}

func notesDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tpl, err := template.ParseFiles(
		"view/templates/base.html",
		"view/templates/pages/detail.html",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	noteId := r.URL.Query().Get("id")
	if noteId == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	
	tpl.ExecuteTemplate(w, "base", map[string]string{"noteName": noteId, "noteContent": "-"})
}

func notesCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl, err := template.ParseFiles(
			"view/templates/base.html",
			"view/templates/create.html",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tpl.ExecuteTemplate(w, "base", nil)
		return
	}
	fmt.Fprint(w, "Notes Create")
}

func main() {
	conf := config.MustLoadConfig()
	// fmt.Println("listening port", conf.ServerPort)

	fmt.Println(conf)

	if err := conf.LoadFromEnv(); err != nil {
		panic(err)
	}
	fmt.Println(conf.ServerHost, conf.ServerPort, conf.SecretKey)
	// mux := http.NewServeMux()

	// staticHandle := http.FileServer(http.Dir("view/static/"))
	// mux.Handle("/static/", http.StripPrefix("/static/", staticHandle))
	
	// mux.HandleFunc("/", homeHandler)
	// mux.HandleFunc("/notes", listNotes)
	// mux.HandleFunc("/notes/detail", notesDetail)
	// mux.HandleFunc("/notes/create", notesCreate)
	
	// http.ListenAndServe(fmt.Sprintf("%s:%s", conf.ServerHost, conf.ServerPort), mux)
}