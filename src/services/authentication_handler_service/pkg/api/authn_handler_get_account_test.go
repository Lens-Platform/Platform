package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	_ "github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/middleware"
)

func TestGetAccountHandler(t *testing.T) {
	var testDataInfo = []struct {
		email                                   string
		responseCode                            int
		errorExpectedToOcurr                    bool
		shouldCreateAndAuthenticateAccountFirst bool
	}{
		{
			fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10)),
			http.StatusOK,
			false,
			true,
		},
		{
			"",
			http.StatusBadRequest,
			true,
			false,
		},
		{
			fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10)),
			http.StatusBadRequest,
			true,
			false,
		},
	}

	for _, data := range testDataInfo {
		var result CreateAccountResponse
		var authRes LoginAccountResponse
		var err error

		// first we create the account
		if data.shouldCreateAndAuthenticateAccountFirst {
			result, err, authRes = createAndLoginAccountTestUtil(t, result, err, data.email, data.errorExpectedToOcurr, authRes)
		}

		// try the update operation
		userId := fmt.Sprint(result.Id)
		opResult, err, rr := GetUserAccountRequestTestUtil(userId, authRes.Token, t)

		if data.errorExpectedToOcurr && (err == nil || (opResult != nil && opResult.Account != nil)) {
			t.Errorf("expected error to occur but none did")
		}

		if !data.errorExpectedToOcurr && (err != nil || (opResult != nil && opResult.Account == nil)) {
			t.Errorf("error was not expected to occur - error %s", err.Error())
		}

		// Check the status code is what we expect.
		if status := rr.Code; status != data.responseCode {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, data.responseCode)
		}

		if opResult.Error != nil && !data.errorExpectedToOcurr {
			t.Errorf("handler returned empty email field. Expected valid response error")
		}

		if opResult != nil && opResult.Account != nil {
			if opResult.Account.Username != data.email || fmt.Sprint(opResult.Account.ID) != userId {
				t.Errorf("invalid operation parameters")
			}
		}
	}
}

func GetUserAccountRequestTestUtil(userId, token string, t *testing.T) (*GetAccountResponse, error, *httptest.ResponseRecorder) {
	var result GetAccountResponse

	req, err := http.NewRequest("GET", "/v1/account/"+userId, nil)
	if err != nil {
		t.Fatal(err)
	}

	req, rr := generateAuthorizedRequest(req, token)
	req = mux.SetURLVars(req, map[string]string{"id": userId})

	srv := NewMockServer()

	handler := http.HandlerFunc(srv.getAccountHandler)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &result)

	return &result, err, rr
}
