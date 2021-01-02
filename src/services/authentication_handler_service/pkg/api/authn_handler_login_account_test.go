package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginAccountHandler(t *testing.T) {
	var testDataInfo = []struct {
		email                        string
		password                     string
		responseCode                 int
		errorExpectedToOcurr         bool
		shouldCreateUserAccountFirst bool
	}{
		{
			GenerateRandomString(5),
			GenerateRandomString(5),
			http.StatusOK,
			false,
			true,
		},
		{
			GenerateRandomString(5),
			GenerateRandomString(5),
			http.StatusInternalServerError,
			true,
			false,
		},
		{
			"",
			"",
			http.StatusBadRequest,
			true,
			false,
		},
		{
			GenerateRandomString(5),
			"",
			http.StatusBadRequest,
			true,
			false,
		},
		{
			"",
			GenerateRandomString(5),
			http.StatusBadRequest,
			true,
			false,
		},
	}

	for _, data := range testDataInfo {
		// first we create the account
		if data.shouldCreateUserAccountFirst {
			result, err, _ := CreateUserAccountRequestTestUtil(data.email, data.password, t)

			if err != nil && !data.errorExpectedToOcurr {
				t.Errorf("obtained error but not expected - %s", err.Error())
			}

			if result.Error != nil && !data.errorExpectedToOcurr {
				t.Errorf("obtained error but not expected - %s", result.Error)
			}
		}

		result, err, rr := LoginUserAccountRequestTestUtil(data.email, data.password, t)
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

		// here we dont test the alternate case because we can have instances in which we return early
		// from the login flow due to some obtained error such as error during the request
		// validation process
		if result.Error != nil && !data.errorExpectedToOcurr {
			t.Errorf("handler returned error")
		}

		if result.Token == "" && !data.errorExpectedToOcurr {
			t.Errorf("null or empty jwt token not expected")
		} else if result.Token != "" && data.errorExpectedToOcurr {
			t.Errorf("expected empty json web token but got a token instead")
		}
	}
}

func LoginUserAccountRequestTestUtil(email, password string, t *testing.T) (LoginAccountResponse, error, *httptest.ResponseRecorder) {
	var result LoginAccountResponse

	reqBody := LoginAccountRequest{
		Email:    email,
		Password: password,
	}

	body, err := createRequestBody(reqBody)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/v1/account/login", body)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv := NewMockServer()
	handler := http.HandlerFunc(srv.loginAccountHandler)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &result)

	return result, err, rr
}
