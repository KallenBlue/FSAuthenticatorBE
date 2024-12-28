package service

import (
	"encoding/json"
	"net/http"
)

func CommonResponse(w http.ResponseWriter, res *JsonResult) {
	w.Header().Set("content-type", "application/json")
	msg, err := json.Marshal(res)
	if err != nil {
		w.Write([]byte("error"))
		return
	}
	w.Write(msg)
}
