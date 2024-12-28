package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wxcloudrun-golang/middlewire/gmail"
)

func GmailCodeHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	w.Header().Set("content-type", "application/json")

	decoder := json.NewDecoder(r.Body)
	body := make(map[string]interface{})
	err := decoder.Decode(&body)
	if err != nil {
		res.Code = 500
		res.ErrorMsg = err.Error()
		CommonResponse(w, res)
		return
	}
	email, ok := body["email"].(string)
	if !ok {
		res.Code = 500
		res.ErrorMsg = "email is required"
		CommonResponse(w, res)
		return
	}
	code, err := gmail.GetEmailCode(email)
	if code == "" {
		res.Code = 200
		res.ErrorMsg = fmt.Sprintf("code is empty, err: %v", err)
		CommonResponse(w, res)
		return
	}
	res.Code = 200
	res.Data = code
	CommonResponse(w, res)
}
