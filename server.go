package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	f_web = flag.String("web", "web", "Path to web directory")
)

func RunServer() {

	slog.Info("Booting up server")
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/submit", handleSubmit)

	// http.Handle("/static", http.FileServer(http.Dir("./web/static/")))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	log.Fatal(http.ListenAndServe(":3030", nil))
}

func handleMain(w http.ResponseWriter, r *http.Request) {
	slog.Debug("Entering handleMain handler")

	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	tmpl.Execute(w, nil)
}

func handleSubmit(w http.ResponseWriter, r *http.Request) {
	slog.Info("Submit handler pinged")
	if r.Method != "POST" {
		slog.Debug("Submit called not using POST request")
		return
	}
	err := r.ParseForm()
	if err != nil {
		slog.Error("Error parsing request form")
		http.Error(w, "Error with parsing form", http.StatusBadRequest)
		return
	}

	submitValues, err := ParseSubmitForm(r.Form)
	if err != nil {
		slog.Error(fmt.Sprintf("%v", err))
		http.Error(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	fmt.Printf("%+v\n", submitValues)

	res := RunPath(submitValues.Attack, submitValues.Path)
	fmt.Println(res)
	odds := res.GetSuccessOdds()
	fmt.Fprintf(w, "<p>Odds of success: %.2f%%</p>", odds)

}

type SubmitReq struct {
	Attack int
	Path   []int
}

func ParseSubmitForm(form url.Values) (*SubmitReq, error) {
	attack_str, ok := form["attack"]
	if !ok {
		err := "Key 'attack' not found in submit form"
		slog.Error(err)
		return nil, fmt.Errorf(err)
	}

	attackVal, err := ParseAttackForm(attack_str)
	if err != nil {
		return nil, err
	}

	path_str, ok := form["path"]
	if !ok {
		err := "Key 'path' not found in submit form"
		slog.Error(err)
		return nil, fmt.Errorf(err)
	}

	pathVal, err := ParsePathForm(path_str)
	if err != nil {
		return nil, err
	}

	sr := &SubmitReq{
		Attack: attackVal,
		Path:   pathVal,
	}

	return sr, nil
}

func ParseAttackForm(attackForm []string) (int, error) {
	if len(attackForm) != 1 {
		return -1, fmt.Errorf("Error parsing attack form: length of list > 1")
	}
	attackStrVal := attackForm[0]
	attackVal, err := strconv.Atoi(attackStrVal)
	if err != nil {
		return -1, fmt.Errorf("Error parsing string as integer: %v", err)
	}
	return attackVal, nil
}

func ParsePathForm(pathForm []string) ([]int, error) {
	var path []int
	if len(pathForm) != 1 {
		return path, fmt.Errorf("Error parsing pathForm: length != 1")
	}
	pathStrVal := pathForm[0]
	commaSplit := strings.Split(pathStrVal, ",")
	for _, val := range commaSplit {
		if val == "" {
			continue
		}
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return path, fmt.Errorf("Error parsing %v as an int", val)
		}
		path = append(path, intVal)
	}

	return path, nil
}
