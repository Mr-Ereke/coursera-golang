package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestCase struct {
	req SearchRequest
	err string
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	limit := r.FormValue("limit")

	switch limit {
	case "1":
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `[{"Id":1,"Name":"Hilda","Age":21,"About":"green","Gender":"female"}]`)
		fallthrough
	case "2":
		w.WriteHeader(http.StatusUnauthorized)
		break
	case "3":
		w.WriteHeader(http.StatusBadRequest)
		break
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{
			req: SearchRequest{
				Limit: -1,
				Offset: 5,
				Query: "",
				OrderField: "id",
				OrderBy: 1,
			},
			err: "limit must be > 0",
		},
		TestCase{
			req: SearchRequest{
				Limit: 1,
				Offset: 5,
				Query: "",
				OrderField: "id",
				OrderBy: 1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit: 2,
				Offset: 5,
				Query: "",
				OrderField: "id",
				OrderBy: 1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit: 3,
				Offset: 5,
				Query: "",
				OrderField: "id",
				OrderBy: 1,
			},
			err: "",
		},
		TestCase{
			req: SearchRequest{
				Limit: 26,
				Offset: -4,
				Query: "",
				OrderField: "id",
				OrderBy: 1,
			},
			err: "",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for _, item := range cases {
		client := &SearchClient{
			AccessToken: "zxc",
			URL: ts.URL,
		}

		_, err := client.FindUsers(item.req)

		if err != nil  {
			//t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}

		//if err == nil {
		//	t.Errorf("[%d] expected error, got nil", caseNum)
		//}
		//
		//if !reflect.DeepEqual(req.Result, result) {
		//	t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
		//}
	}

	ts.Close()
}