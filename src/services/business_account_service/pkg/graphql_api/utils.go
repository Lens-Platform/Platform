package graphql_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"
	"unsafe"

	"github.com/gorilla/mux"
)

type CreateAccountDtxRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountDtxResponse struct {
	Error error  `json:"error"`
	Id    uint32 `json:"id"`
}

type LockAccountDtxResponse struct {
	Error error `json:"error"`
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func HandleErrorsIfPresent(ID uint32) (*int, error) {
	if ID == 0 {
		return nil, errors.New("failed to obtain business account id value")
	}

	retVal, err := strconv.Atoi(fmt.Sprint(ID))
	if err != nil {
		return nil, err
	}

	return &retVal, nil
}

func CreateAccountInAuthServiceAndGetAuthToken(t *testing.T, email, password string) (string, uint32) {
	var result CreateAccountDtxResponse
	reqBody := CreateAccountDtxRequest{
		Email:    email,
		Password: password,
	}

	body, err := CreateRequestBody(reqBody)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "http://localhost:9898/v1/account/create", body)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	// authenticate account
	return AuthenticateAccountAndReturnJwtToken(t, email, password), result.Id
}

func AuthenticateAccountAndReturnJwtToken(t *testing.T, email, password string) string {
	type AuthResult struct {
		Error error  `json:"error"`
		Token string `json:"token"`
	}

	reqBody := CreateAccountDtxRequest{
		Email:    email,
		Password: password,
	}

	body, err := CreateRequestBody(reqBody)
	if err != nil {
		t.Fatal(err)
	}

	var result AuthResult
	req, err := http.NewRequest("POST", "http://localhost:9898/v1/account/login", body)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	return result.Token
}

func LockAccountInAuthService(t *testing.T, authnId uint32, token string) error {
	var response LockAccountDtxResponse
	id := fmt.Sprint(authnId)
	req, err := http.NewRequest("POST", "http://localhost:9898/v1/account/lock/"+id, nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		t.Fatal(err)
	}

	return response.Error
}

func CreateRequestBody(body interface{}) (*bytes.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b), nil
}
