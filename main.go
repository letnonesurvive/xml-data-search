package main

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type XMLClient struct {
	Id            int    `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type Clients struct {
	Clients []XMLClient `xml:"row"`
}

func SortAsc(sortType string, responceUsers []User) {
	switch sortType {
	case "Id":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Id < responceUsers[j].Id
		})
	case "Age":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Age < responceUsers[j].Age
		})
	default:
		fallthrough
	case "Name":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Name < responceUsers[j].Name
		})
	}
}

func SortDesc(sortType string, responceUsers []User) {
	switch sortType {
	case "Id":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Id > responceUsers[j].Id
		})
	case "Age":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Age > responceUsers[j].Age
		})
	default:
		fallthrough
	case "Name":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Name > responceUsers[j].Name
		})
	}
}

func SortClients(sortType string, sortOrder int, responceUsers []User) {

	switch sortOrder {
	case OrderByAsIs:
		return
	case OrderByAsc:
		SortAsc(sortType, responceUsers)
	case OrderByDesc:
		SortDesc(sortType, responceUsers)
	default:
		fmt.Println("Wrong sort order")
		return
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("dataset.xml")
	if err != nil {
		fmt.Println("Failed to open dataset.xml")
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var clients Clients
	decoder.Decode(&clients)

	query := r.FormValue("query")
	if len(query) == 0 {
		fmt.Println("query is empty")
		return
	}
	var responceUsers []User
	for _, client := range clients.Clients {
		Name := client.FirstName + " " + client.LastName
		if strings.Contains(Name, query) || strings.Contains(client.About, query) {
			var user User
			user.About = client.About
			user.Age = client.Age
			user.Gender = client.Gender
			user.Id = client.Id
			user.Name = Name
			responceUsers = append(responceUsers, user)
		}
	}

	orderField := r.FormValue("order_field")
	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		fmt.Println("Wrong convertion order_by to int")
	}
	SortClients(orderField, orderBy, responceUsers)

	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		fmt.Println("Wrong convertion limit to int")
	}
	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		fmt.Println("Wrong convertion offset to int")
	}

	if offset < len(responceUsers) && offset > 0 {
		responceUsers = responceUsers[offset:]
	} else {
		fmt.Println("Wrong offset")
	}

	if limit < len(responceUsers) && limit > 0 {
		fmt.Println(responceUsers[:limit])
	} else {
		fmt.Println("Wrong limit")
	}

	fmt.Println(responceUsers)
}

func runServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", SearchServer)

	server := http.Server{
		Addr:        ":8080",
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
	}

	server.ListenAndServe()
}

func main() {
	runServer()
}
