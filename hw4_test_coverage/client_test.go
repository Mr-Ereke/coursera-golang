package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
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
	Gender     string   `xml:"gender"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	offset, err := strconv.Atoi(r.FormValue("offset"))
	limit, err := strconv.Atoi(r.FormValue("limit"))
	//orderBy, err := strconv.Atoi(r.FormValue("order_by"))

	if err != nil {
		panic(err)
	}

	users := ParseXML()

	var filterUsers []User

	if query != "" {
		for _, user := range users {
			if (strings.Contains(user.Name, query) || strings.Contains(user.About, query)) {
				filterUsers = append(filterUsers, user)
			}
		}
	}

	//var orderFieldName string

	switch orderField := r.FormValue("order_field"); orderField {
	case "Id":
		//orderFieldName = "Id"
		sort.Slice(filterUsers[:], func(i, j int) bool {
			return filterUsers[i].Id < filterUsers[j].Id
		})
		break
	case "Age":
		//orderFieldName = "Age"
		sort.Slice(filterUsers[:], func(i, j int) bool {
			return filterUsers[i].Age < filterUsers[j].Age
		})
		break
	case "Name":
		//orderFieldName = "Name"
		break
	case "":
		//orderFieldName = "Name"
		break
	default:
		errors.New(ErrorBadOrderField)
	}

	sliceLimit := 0

	if limit <= len(filterUsers) {
		sliceLimit = limit
	} else {
		sliceLimit = len(filterUsers)
	}

	limitedUsers := filterUsers[offset:sliceLimit]

	var response []string

	for _, user := range limitedUsers {
		if jsn, err := json.Marshal(user); err == nil {
			response = append(response, string(jsn))
		} else {
			log.Panic(err)
		}
	}

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "[" + strings.Join(response, ",") + "]")

	//switch limit {
	//case "1":
	//	w.WriteHeader(http.StatusOK)
	//	io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
	//	fallthrough
	//case "2":
	//	w.WriteHeader(http.StatusUnauthorized)
	//	io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
	//	breakoffset
	//case "3":
	//	w.WriteHeader(http.StatusBadRequest)
	//	io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
	//	break
	//case "4":
	//	w.WriteHeader(http.StatusBadRequest)
	//	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	//	break
	//default:
	//	w.WriteHeader(http.StatusInternalServerError)
	//	io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
	//}
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
		newUser.Gender = rows.Row[i].Gender
		users = append(users, newUser)
	}

	return users
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{
			req: SearchRequest{
				Limit:      30,
				Offset:     0,
				Query:      "minim",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      -1,
				Offset:     0,
				Query:      "minim",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      1,
				Offset:     -1,
				Query:      "minim",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      0,
				Offset:     0,
				Query:      "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum",
				OrderField: "Id",
				OrderBy:    1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum",
				OrderField: "Id",
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
