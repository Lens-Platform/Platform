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

func TestLockAccountHandler(t *testing.T) {
	var testDataInfo = []struct {
		newEmail                                string
		responseCode                            int
		errorExpectedToOcurr                    bool
		shouldCreateAndAuthenticateAccountFirst bool
	}{
		{
			// test case where we have a valid account and lock it
			fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10)),
			http.StatusOK,
			false,
			true,
		},
		{
			// test case where we have an invalid account and cant lock it
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
			result, err, authRes = createAndLoginAccountTestUtil(t, result, err, data.newEmail, data.errorExpectedToOcurr, authRes)
		}

		// try the lock operation
		userId := fmt.Sprint(result.Id)
		opResult, err, rr := LockUserAccountRequestTestUtil(userId, authRes.Token, t)

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
		if opResult.Error != nil && !data.errorExpectedToOcurr {
			t.Errorf("handler returned empty email field. Expected valid response error")
		}
	}
}

func LockUserAccountRequestTestUtil(userId, token string, t *testing.T) (LockAccountResponse, error, *httptest.ResponseRecorder) {
	var result LockAccountResponse

	req, err := http.NewRequest("POST", "/v1/account/lock/"+userId, nil)
	if err != nil {
		t.Fatal(err)
	}

	req, rr := generateAuthorizedRequest(req, token)
	req = mux.SetURLVars(req, map[string]string{"id": userId})

	srv := NewMockServer()
	handler := http.HandlerFunc(srv.lockAccountHandler)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &result)

	return result, err, rr
}
