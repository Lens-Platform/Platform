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

func TestDeleteAccountHandler(t *testing.T) {
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
			fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10)),
			http.StatusBadRequest,
			true,
			false,
		},
		{
			"",
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

		// try the delete operation
		userId := fmt.Sprint(result.Id)
		opResult, err, rr := DeleteUserAccountRequestTestUtil(userId, authRes.Token, t)

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

		if opResult.Error != nil && !data.errorExpectedToOcurr {
			t.Errorf("handler returned empty email field. Expected valid response error")
		}
	}
}

func DeleteUserAccountRequestTestUtil(userId, token string, t *testing.T) (DeleteAccountResponse, error, *httptest.ResponseRecorder) {
	var result DeleteAccountResponse

	req, err := http.NewRequest("DELETE", "/v1/account/delete/"+userId, nil)
	if err != nil {
		t.Fatal(err)
	}

	req, rr := generateAuthorizedRequest(req, token)
	req = mux.SetURLVars(req, map[string]string{"id": userId})

	srv := NewMockServer()

	handler := http.HandlerFunc(srv.deleteAccountHandler)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &result)

	return result, err, rr
}
