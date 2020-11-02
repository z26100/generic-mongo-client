package mongo

import (
	"log"
	"net/http"
)

func check(condition func() bool, w http.ResponseWriter) bool {
	if condition() {
		http.Error(w, "BadRequest", http.StatusBadRequest)
		return true
	}
	return false
}

func checkError(err error, w http.ResponseWriter) bool {
	return check(func() bool {
		if err != nil {
			log.Println(err)
		}
		return err != nil
	}, w)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "*")
}
