package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAccountHandler(t *testing.T) {
	email := fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(17))
	password := fmt.Sprintf("test_password_%s", GenerateRandomString(17))

	var testDataInfo = []struct {
		email                string
		password             string
		responseCode         int
		errorExpectedToOcurr bool
	}{
		{
			email,
			password,
			http.StatusOK,
			false,
		},
		{
			// case we have duplicates
			email,
			password,
			http.StatusInternalServerError,
			true,
		},
		{
			"",
			"",
			http.StatusBadRequest,
			true,
		},
		{
			GenerateRandomString(15),
			"",
			http.StatusBadRequest,
			true,
		},
		{
			"",
			GenerateRandomString(15),
			http.StatusBadRequest,
			true,
		},
	}

	for _, data := range testDataInfo {
		result, err, rr := CreateUserAccountRequestTestUtil(data.email, data.password, t)
		if data.errorExpectedToOcurr && err == nil {
			t.Errorf("expected error to occur but none did")
		}

		if !data.errorExpectedToOcurr && err != nil {
			t.Errorf("error was not expected to occur - error %s", err.Error())
		}

		// Check the status code is what we expect.
		if status := rr.Code; status != data.responseCode {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, data.responseCode)
		}

		if result.Error != nil && !data.errorExpectedToOcurr {
			t.Errorf("handler returned error")
		}
	}
}

func CreateUserAccountRequestTestUtil(email, password string, t *testing.T) (CreateAccountResponse, error, *httptest.ResponseRecorder) {
	var result CreateAccountResponse

	reqBody := CreateAccountRequest{
		Email:    email,
		Password: password,
	}

	body, err := createRequestBody(reqBody)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "//v1/account/create", body)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv := NewMockServer()
	handler := http.HandlerFunc(srv.createAccountHandler)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &result)

	return result, err, rr
}
