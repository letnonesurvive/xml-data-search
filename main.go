package main

import (
	"encoding/xml"
	"fmt"
	"os"
)

type XMLClient struct {
	Id            int    `xml:"id"`
	Guid          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           uint8  `xml:"age"`
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

func SearchServer(req SearchRequest) {
	file, err := os.Open("dataset.xml")
	if err != nil {
		fmt.Println("Failed to open dataset.xml")
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)

	var clients Clients
	decoder.Decode(&clients)

	for client := range clients.Clients {
		fmt.Println(client)
	}
}

func main() {
	var req SearchRequest
	req.Query = "Boyd"
	SearchServer(req)
}
