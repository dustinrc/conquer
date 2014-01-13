package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/sessions"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
)

var (
	username    = flag.String("username", "whiskey", "the username for login authentication")
	password    = flag.String("password", "hotel", "the password for login authentication")
	store       = sessions.NewCookieStore([]byte("shh, don't tell anyone"))
	sessionName = "default"
	loginStatus = "loggedIn"

	configList = NewConfigList()

	bigFilterPath = "large"
	// issues uploading configs with a greater number of entries, using a smaller number
	// to still show the filtering capability
	bigFilterNumber = 100
)

// jsonresponse is the JSON body for all responses, ConfigList is appended as necessary
type jsonResponse struct {
	Id    int    `json:"id"`
	Error string `json:"error"`
	*ConfigList
}

// String will pretty print the jsonResponse
func (jr jsonResponse) String() string {
	b, _ := json.MarshalIndent(jr, "", "    ")
	return string(b)
}

// jsonrequest is the JSON body/data for all ConfigList based actions
type jsonRequest struct {
	Id int `json:"id"`
	*ConfigList
}

// jsonAuth is the JSON body solely forget login action
type jsonAuth struct {
	Id   int    `json:"id"`
	User string `json:"username"`
	Pass string `json:"password"`
}

// getConfigs retrieves the ConfigList and filters/sorts as controlled
// by the extended URL
func getConfigs(w http.ResponseWriter, id int, splits []string) {
	var filter, order []string
	if len(splits) > 0 {
		if splits[0] == bigFilterPath {
			filter = splits[:1]
			order = splits[1:]
		} else {
			order = splits
		}
		configList.OrderBy(order)
		sort.Sort(configList)
		configList.OrderBy([]string{})
	}

	var resp jsonResponse
	resp.Id = id
	if len(filter) > 0 {
		filteredList := NewConfigList()
		for _, config := range configList.Configs {
			if len(config.Vars) > bigFilterNumber {
				filteredList.add(config)
			}
		}
		resp.ConfigList = filteredList
	} else {
		resp.ConfigList = configList
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, resp)
	configList.OrderBy([]string{})
}

// addConfigs append the provided ConfigList to the main.  A list
// of redundant configs (by the unique variable) is returned.
func addConfigs(w http.ResponseWriter, id int, list *ConfigList) {
	failedList := configList.Add(list)
	resp := jsonResponse{Id: id, ConfigList: failedList}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, resp)
}

// deleteconfigs remove Configs from the main list that have matching
// unique variables with those in the given ConfigList.  Non-matching
// items in the given list will be returned.
func deleteConfigs(w http.ResponseWriter, id int, list *ConfigList) {
	failedList := configList.Delete(list)
	resp := jsonResponse{Id: id, ConfigList: failedList}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, resp)
}

// badMethod indicates an unsupported HTTP method for the URL
func badMethod(w http.ResponseWriter, id int) {
	resp := jsonResponse{Id: id, Error: "bad method"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintln(w, resp)
}

// badRequest indicates a malformed request, most likely incorrect JSON
func badRequest(w http.ResponseWriter, id int) {
	resp := jsonResponse{Id: id, Error: "bad request"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintln(w, resp)
}

// ok indicates a successful login
func ok(w http.ResponseWriter, id int) {
	resp := jsonResponse{Id: id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, resp)
}

// unauthorized indicates the user is not authenticated for the current session
func unauthorized(w http.ResponseWriter, id int) {
	resp := jsonResponse{Id: id, Error: "unauthorized"}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization Required\"")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintln(w, resp)
}

// loginHandler allows the user to authenticate the current session
func loginHandler(w http.ResponseWriter, r *http.Request) {
	auth := jsonAuth{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&auth)
	if err != nil && err != io.EOF {
		badRequest(w, auth.Id)
		return
	}

	if r.Method != "POST" {
		badMethod(w, auth.Id)
		return
	}

	if auth.User != *username || auth.Pass != *password {
		unauthorized(w, auth.Id)
		return
	}

	s, _ := store.Get(r, sessionName)
	s.Values[loginStatus] = true
	s.Save(r, w)
	ok(w, auth.Id)
}

// logoutHandler allows the user to deactivate the current session
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	req := jsonRequest{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil && err != io.EOF {
		badRequest(w, req.Id)
		return
	}

	if r.Method != "POST" {
		badMethod(w, req.Id)
		return
	}

	s, _ := store.Get(r, sessionName)
	s.Values[loginStatus] = false
	s.Save(r, w)
	ok(w, req.Id)
}

// configHandler ensures the user is logged in and then takes the
// appropriate action on the ConfigList based on the HTTP method given
func configsHandler(w http.ResponseWriter, r *http.Request) {
	req := jsonRequest{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil && err != io.EOF {
		badRequest(w, req.Id)
		return
	}

	s, err := store.Get(r, sessionName)
	if err != nil {
		unauthorized(w, req.Id)
		return
	}

	if loggedIn, ok := s.Values[loginStatus]; !ok {
		unauthorized(w, req.Id)
		return
	} else if loggedIn.(bool) {

		switch r.Method {
		case "GET":
			extendedPath := r.URL.Path[len("/configs/"):]
			extendedPathSplits := strings.Split(extendedPath, "/")
			getConfigs(w, req.Id, extendedPathSplits)
		case "POST":
			addConfigs(w, req.Id, req.ConfigList)
		case "DELETE":
			deleteConfigs(w, req.Id, req.ConfigList)
		default:
			badMethod(w, req.Id)
		}
		return

	} else {
		unauthorized(w, req.Id)
	}
}

// main parses the command line args, sets up the server routes,
// and runs the server
func main() {
	flag.Parse()

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/configs/", configsHandler)

	fmt.Println()
	fmt.Println("Listening on 0.0.0.0:8080")
	fmt.Println("  /login                                 [POST]")
	fmt.Println("  /logout                                [POST]")
	fmt.Println("  /configs/                              [POST,DELETE]")
	fmt.Println("  /configs/{name,hostname,port,username} [GET]")
	fmt.Println("  /configs/large                         [GET]")
	fmt.Println("The username/password is ", *username, "/", *password)
	fmt.Println()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
