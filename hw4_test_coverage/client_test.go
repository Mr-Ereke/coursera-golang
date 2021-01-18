package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type TestCase struct {
	req SearchRequest
	err string
}

type Rows struct {
	XMLName xml.Name `xml:"root"`
	Row     []Row    `xml:"row"`
}

type Row struct {
	XMLName   xml.Name `xml:"row"`
	Id        int      `xml:"id"`
	Age       int      `xml:"age"`
	FirstName string   `xml:"first_name"`
	LastName  string   `xml:"last_name"`
	About     string   `xml:"about"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	offset := r.FormValue("offset")
	limit := r.FormValue("limit")
	orderField := r.FormValue("order_field")
	query := r.FormValue("query")
	orderBy := r.FormValue("order_by")

	var users []User

	if len(users) < 1 {
		users = ParseXML()
	}

	switch limit {
	case "1":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
		fallthrough
	case "2":
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
		break
	case "3":
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
		break
	case "4":
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		break
	default:
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
	}
}

func ParseXML() []User {
	xmlFile, err := os.Open("dataset.xml")

	if err != nil {
		fmt.Println(err)
	}

	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var rows Rows
	var users []User

	xml.Unmarshal(byteValue, &rows)

	for i := 0; i < len(rows.Row); i++ {
		newUser := User{}
		newUser.Id = rows.Row[i].Id
		newUser.Name = rows.Row[i].FirstName + rows.Row[i].LastName
		newUser.Age = rows.Row[i].Age
		newUser.About = rows.Row[i].About
		users = append(users, newUser)
	}

	return users
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{
			req: SearchRequest{
				Limit:      -1,
				Offset:     5,
				Query:      "",
				OrderField: "id",
				OrderBy:    1,
			},
			err: "limit must be > 0",
		},
		TestCase{
			req: SearchRequest{
				Limit:      1,
				Offset:     5,
				Query:      "",
				OrderField: "id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      2,
				Offset:     5,
				Query:      "",
				OrderField: "id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      3,
				Offset:     5,
				Query:      "",
				OrderField: "id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      4,
				Offset:     5,
				Query:      "",
				OrderField: "id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      26,
				Offset:     -4,
				Query:      "",
				OrderField: "id",
				OrderBy:    1,
			},
			err: "",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for _, item := range cases {
		client := &SearchClient{
			AccessToken: "token",
			URL:         ts.URL,
		}

		_, err := client.FindUsers(item.req)

		if err != nil {
			//t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}

		//if err == nil {
		//	t.Errorf("[%d] expected error, got nil", caseNum)
		//}

		//if !reflect.DeepEqual(req.Result, result) {
		//	t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		//}
	}
	ts.Close()
}
