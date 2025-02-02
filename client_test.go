package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
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
	case "":
		fallthrough
	case "Name":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Name < responceUsers[j].Name
		})
	default:
		fmt.Println("Wrong sort_type")
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
	case "":
		fallthrough
	case "Name":
		sort.Slice(responceUsers, func(i, j int) bool {
			return responceUsers[i].Name > responceUsers[j].Name
		})
	default:
		fmt.Println("Wrong sort_type")
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
	orderByStr := r.FormValue("order_by")
	orderBy, err := strconv.Atoi(orderByStr)
	if err != nil {
		http.Error(w, "Wrong convertion order_by to int", http.StatusBadRequest)
		return
	}
	SortClients(orderField, orderBy, responceUsers)

	limitStr := r.FormValue("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		http.Error(w, "Wrong convertion limit to int", http.StatusBadRequest)
		return
	}
	offsetStr := r.FormValue("offset")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		http.Error(w, "Wrong convertion offset to int", http.StatusBadRequest)
		return
	}

	if offset < len(responceUsers) && offset >= 0 {
		responceUsers = responceUsers[offset:]
	} else {
		http.Error(w, "Wrong offset", http.StatusBadRequest)
		return
	}

	if limit < len(responceUsers) {
		responceUsers = responceUsers[:limit]
	} else if limit < 0 {
		http.Error(w, "Wrong limit", http.StatusBadRequest)
		return
	}

	w.WriteHeader(200)
	bytes, err := json.Marshal(responceUsers)
	if err != nil {
		fmt.Println("Unable to marshall users to json")
	}
	w.Write(bytes)
}

type TestCase struct {
	Request  SearchRequest
	Response SearchResponse
}

func TestRequestSingleUser(t *testing.T) {
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "Hilda",
				Limit: 1,
			},
			Response: SearchResponse{
				Users: []User{
					{
						Id:     1,
						Name:   "Hilda Mayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(SearchServer))

	for testNum, testCase := range testCases {
		var client SearchClient
		client.URL = server.URL
		responsePoint, err := client.FindUsers(testCase.Request)
		if err != nil {
			t.Errorf("[%d] unexpected error: %#v", testNum, err)
		}
		response := *responsePoint
		if !reflect.DeepEqual(response, testCase.Response) { // прикольная штука, надо про нее написать
			t.Errorf("[%d] wrong result, expected %#v, got %#v", testNum, testCase.Response, response)
		}
	}
}
