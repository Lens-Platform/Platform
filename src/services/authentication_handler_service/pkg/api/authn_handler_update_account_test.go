package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/middleware"
	_ "github.com/BlackspaceInc/BlackspacePlatform/src/services/authentication_handler_service/pkg/middleware"
)

func TestUpdateAccountHandler(t *testing.T) {
	oldEmail := fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10))
	newEmailSameAsOld := oldEmail

	var testDataInfo = []struct {
		oldEmail                                string
		newEmail                                string
		responseCode                            int
		errorExpectedToOcurr                    bool
		shouldCreateAndAuthenticateAccountFirst bool
	}{
		{
			fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10)),
			fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10)),
			http.StatusOK,
			false,
			true,
		},
		{
			fmt.Sprintf("test_%s@gmail.com", GenerateRandomString(10)),
			"",
			http.StatusBadRequest,
			true,
			false,
		},
		{
			oldEmail,
			newEmailSameAsOld,
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
			result, err, authRes = createAndLoginAccountTestUtil(t, result, err, data.oldEmail, data.errorExpectedToOcurr, authRes)
		}

		// try the update operation
		userId := fmt.Sprint(result.Id)
		updateResult, err, rr := UpdateUserAccountRequestTestUtil(data.newEmail, userId, authRes.Token, t)

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
		if updateResult.Error != nil && !data.errorExpectedToOcurr {
			t.Errorf("handler returned empty email field. Expected valid response error")
		}
	}
}

func UpdateUserAccountRequestTestUtil(email, userId, token string, t *testing.T) (UpdateAccountResponse, error, *httptest.ResponseRecorder) {
	var result UpdateAccountResponse

	reqBody := UpdateAccountRequest{
		Email: email,
	}

	body, err := createRequestBody(reqBody)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/v1/account/update/"+userId, body)
	if err != nil {
		t.Fatal(err)
	}

	req, rr := generateAuthorizedRequest(req, token)
	req = mux.SetURLVars(req, map[string]string{"id": userId})

	srv := NewMockServer()

	handler := http.HandlerFunc(srv.updateAccountHandler)

	handler.ServeHTTP(rr, req)

	err = json.Unmarshal(rr.Body.Bytes(), &result)

	return result, err, rr
}

func generateAuthorizedRequest(req *http.Request, token string) (*http.Request, *httptest.ResponseRecorder) {
	// add jwt token to header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	rr := httptest.NewRecorder()
	ctx := context.Background()

	if token != "" {
		ctx = context.WithValue(ctx, middleware.UserCtxKey, token)
	}

	req = req.WithContext(ctx)
	return req, rr
}

func createAndLoginAccountTestUtil(t *testing.T, result CreateAccountResponse, err error, email string, errorExpectedToOcurr bool,
	authRes LoginAccountResponse) (CreateAccountResponse, error, LoginAccountResponse) {
	password := GenerateRandomString(10)
	result, err, _ = CreateUserAccountRequestTestUtil(email, password, t)

	if err != nil && !errorExpectedToOcurr {
		t.Errorf("obtained error but not expected - %s", err.Error())
	}

	if result.Id == 0 && !errorExpectedToOcurr {
		t.Errorf("obtained error since id is 0 but not expected - %s", result.Error)
	}

	authRes, err, _ = LoginUserAccountRequestTestUtil(email, password, t)
	if errorExpectedToOcurr && err == nil {
		t.Errorf("expected error to occur but none did")
	}

	if !errorExpectedToOcurr && err != nil {
		t.Errorf("error was not expected to occur - error %s", err.Error())
	}

	if authRes.Token == "" && !errorExpectedToOcurr {
		t.Errorf("null or empty jwt token not expected")
	} else if authRes.Token != "" && errorExpectedToOcurr {
		t.Errorf("expected empty json web token but got a token instead")
	}
	return result, err, authRes
}
