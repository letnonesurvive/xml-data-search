package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
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
	AcessToken := r.Header[http.CanonicalHeaderKey("AccessToken")]

	if len(AcessToken) == 0 || AcessToken[0] == "" {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	file, err := os.Open("dataset.xml")
	if err != nil { // to test this case need to create a Handler which has any other .xml file which can't be open
		http.Error(w, "Can't open dataset.xml", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var clients Clients
	decoder.Decode(&clients)

	query := r.FormValue("query")

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

	if offset < len(responceUsers)+1 && offset >= 0 {
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
	Response *SearchResponse
}

func RunTest(t *testing.T, testCases []TestCase) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	server.Config.ReadTimeout = 1000 * time.Second
	server.Config.WriteTimeout = 1000 * time.Second

	for testNum, testCase := range testCases {
		var client SearchClient
		client.AccessToken = "TestToken"
		client.URL = server.URL
		response, err := client.FindUsers(testCase.Request)
		if err != nil {
			t.Errorf("[%d] unexpected error: %#v", testNum, err)
		}
		if !reflect.DeepEqual(response, testCase.Response) { // прикольная штука, надо про нее написать
			t.Errorf("[%d] wrong result, expected \n %#v, got \n %#v", testNum, testCase.Response, response)
		}
	}
}

func TestRequestSingleUser(t *testing.T) {
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "Hilda",
				Limit: 1,
			},
			Response: &SearchResponse{
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

	RunTest(t, testCases)
}

var MultipleUsersCupidatat = []User{
	{
		Id:     0,
		Name:   "Boyd Wolf",
		Age:    22,
		About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
		Gender: "male",
	},
	{
		Id:     1,
		Name:   "Hilda Mayer",
		Age:    21,
		About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
		Gender: "female",
	},
	{
		Id:     5,
		Name:   "Beulah Stark",
		Age:    30,
		About:  "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n",
		Gender: "female",
	},
	{
		Id:     6,
		Name:   "Jennings Mays",
		Age:    39,
		About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
		Gender: "male",
	},
}

var MultipleUsersSearchResponce = SearchResponse{
	Users:    MultipleUsersCupidatat,
	NextPage: true,
}

func TestRequestMultipleUsers(t *testing.T) {
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "cupidatat",
				Limit: 4,
				//Limit :5 shoud work as with 4?
			},
			Response: &MultipleUsersSearchResponce,
		},
		{
			Request: SearchRequest{
				Query:      "cupidatat",
				Limit:      4,
				OrderBy:    OrderByAsIs,
				OrderField: "Id",
				//Limit :5 shoud work as with 4?
			},
			Response: &MultipleUsersSearchResponce,
		},
	}
	RunTest(t, testCases)
}

func TestRequestSortUsers(t *testing.T) {
	var MultipleUsersCupidatatReversed = slices.Clone(MultipleUsersCupidatat)
	slices.Reverse(MultipleUsersCupidatatReversed)
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "cupidatat",
				Limit: 4,
			},
			Response: &MultipleUsersSearchResponce,
		},
		{
			Request: SearchRequest{
				Query:      "cupidatat",
				Limit:      4,
				OrderBy:    OrderByAsIs,
				OrderField: "Id",
			},
			Response: &MultipleUsersSearchResponce,
		},
		{
			Request: SearchRequest{
				Query:      "cupidatat",
				Limit:      4,
				OrderBy:    OrderByAsc,
				OrderField: "Id",
			},
			Response: &MultipleUsersSearchResponce,
		},
	}
	RunTest(t, testCases)
}

var EmptyQuerySearchResponce = SearchResponse{
	Users: []User{
		{
			Id:     0,
			Name:   "Boyd Wolf",
			Age:    22,
			About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
			Gender: "male",
		},
		{
			Id:     1,
			Name:   "Hilda Mayer",
			Age:    21,
			About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
			Gender: "female",
		},
		{
			Id:     2,
			Name:   "Brooks Aguilar",
			Age:    25,
			About:  "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n",
			Gender: "male",
		},
		{
			Id:     3,
			Name:   "Everett Dillard",
			Age:    27,
			About:  "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
			Gender: "male",
		},
		{
			Id:     4,
			Name:   "Owen Lynn",
			Age:    30,
			About:  "Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n",
			Gender: "male",
		},
		{
			Id:     5,
			Name:   "Beulah Stark",
			Age:    30,
			About:  "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n",
			Gender: "female",
		},
		{
			Id:     6,
			Name:   "Jennings Mays",
			Age:    39,
			About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
			Gender: "male",
		},
		{
			Id:     7,
			Name:   "Leann Travis",
			Age:    34,
			About:  "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.\n",
			Gender: "female",
		},
		{
			Id:     8,
			Name:   "Glenn Jordan",
			Age:    29,
			About:  "Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n",
			Gender: "male",
		},
		{
			Id:     9,
			Name:   "Rose Carney",
			Age:    36,
			About:  "Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n",
			Gender: "female",
		},
		{
			Id:     10,
			Name:   "Henderson Maxwell",
			Age:    30,
			About:  "Ex et excepteur anim in eiusmod. Cupidatat sunt aliquip exercitation velit minim aliqua ad ipsum cillum dolor do sit dolore cillum. Exercitation eu in ex qui voluptate fugiat amet.\n",
			Gender: "male",
		},
	},
	NextPage: true,
}

func TestEmptyQuery(t *testing.T) {
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "",
				Limit: 11,
			},
			Response: &EmptyQuerySearchResponce,
		},
	}
	RunTest(t, testCases)

	testCasesSize := []TestCase{
		{
			Request: SearchRequest{
				Query: "",
				Limit: 35,
			},
			Response: &SearchResponse{
				Users:    make([]User, 25), // maximum limit is 25
				NextPage: true,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	for testNum, testCase := range testCasesSize {
		var client SearchClient
		client.URL = server.URL
		client.AccessToken = "TestToken"
		response, err := client.FindUsers(testCase.Request)
		if err != nil {
			t.Errorf("[%d] unexpected error: %#v", testNum, err)
		}
		if len(response.Users) != len(testCase.Response.Users) {
			t.Errorf("[%d] wrong result, expected \n %#v, got \n %#v", testNum, testCase.Response, response)
		}
	}
}

func TestFindUserByName(t *testing.T) {

	searchResponce := SearchResponse{
		Users: []User{
			{
				Id:     0,
				Name:   "Boyd Wolf",
				Age:    22,
				About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
				Gender: "male",
			},
		},
		NextPage: false,
	}

	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "Boyd Wolf",
				Limit: 1,
			},
			Response: &searchResponce,
		},
		{
			Request: SearchRequest{
				Query: "Boyd",
				Limit: 1,
			},
			Response: &searchResponce,
		},
		{
			Request: SearchRequest{
				Query: "Wolf",
				Limit: 1,
			},
			Response: &searchResponce,
		},
	}
	RunTest(t, testCases)
}

func TestEmptyResponce(t *testing.T) {
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "This string doesn't contain in dataset.xml file",
			},
			Response: &SearchResponse{
				Users:    nil,
				NextPage: false,
			},
		},
	}
	RunTest(t, testCases)
}

func TestWrongRequestParams(t *testing.T) {
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "Some query",
				Limit: -1,
			},
			Response: nil,
		},
		{
			Request: SearchRequest{
				Query:  "Some query",
				Offset: -1,
			},
			Response: nil,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(SearchServer))

	for testNum, testCase := range testCases {
		var client SearchClient
		client.URL = server.URL
		client.AccessToken = "TestToken"
		response, err := client.FindUsers(testCase.Request)
		if !((err.Error() == "limit must be > 0") || (err.Error() == "offset must be > 0")) {
			t.Errorf("[%d] unexpected error: %#v", testNum, err)
		}
		if response != testCase.Response {
			t.Errorf("[%d] wrong result, expected \n %#v, got \n %#v", testNum, testCase.Response, response)
		}
	}
}

func TestBadAccessToken(t *testing.T) {
	testCases := []TestCase{
		{
			Request: SearchRequest{
				Query: "Boyd",
				Limit: 1,
			},
			Response: nil,
		},
	}

	server := httptest.NewServer(http.HandlerFunc(SearchServer))

	for testNum, testCase := range testCases {
		var client SearchClient
		client.URL = server.URL
		response, err := client.FindUsers(testCase.Request)
		if err.Error() != "Bad AccessToken" {
			t.Errorf("[%d] unexpected error: %#v", testNum, err)
		}
		if response != testCase.Response {
			t.Errorf("[%d] wrong result, expected \n %#v, got \n %#v", testNum, testCase.Response, response)
		}
	}
}
