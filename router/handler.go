package router

import (
	"database/sql"
	"fmt"
	"github.com/angelhack2019/lib/utility"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	connectionString = "host=localhost user=default password=default dbname=dough_you sslmode=disable port=5432"

	errStrUnableToReachDB = "unable to connect to db"
)

var (
	postgresDB *sql.DB
)

func refreshDBConnection() error {
	if postgresDB == nil {
		var err error
		postgresDB, err = sql.Open("postgres", connectionString)
		if err != nil {
			return err
		}
	}

	if err := postgresDB.Ping(); err != nil {
		_ = postgresDB.Close()
		postgresDB = nil
		return err
	}

	return nil
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func getFood(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Get Food")
	//todos := Todos{
	//	Todo{Name: "Write presentation"},
	//	Todo{Name: "Host meetup"},
	//}
	//
	//if err := json.NewEncoder(w).Encode(todos); err != nil {
	//	panic(err)
	//}
}

func getFoods(w http.ResponseWriter, r *http.Request) {
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}
	tags := r.FormValue("tags")

	if strings.TrimSpace(tags) == "" {

	}

	command := `
				INSERT INTO dough_you.foods(
					uuid, pic_url, exp_date, created_date 
				) VALUES($1, $2, $3, $4)
				`
	_, err := postgresDB.Exec(command, uuid.New().String(), picURL, time.Unix(int64(t), 0), time.Now().UTC())

	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "unable to get food")
		return
	}
	utility.Respond(w, http.StatusOK, "OK")
}

func shareFood(w http.ResponseWriter, r *http.Request) {
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}
	picURL := r.FormValue("pic_url")
	expDate := r.FormValue("exp_date")
	t, err := strconv.Atoi(expDate)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "wrong expiration date")
		return
	}
	command := `
				INSERT INTO dough_you.foods(
					uuid, pic_url, exp_date, created_date 
				) VALUES($1, $2, $3, $4)
				`
	_, err = postgresDB.Exec(command, uuid.New().String(), picURL, time.Unix(int64(t), 0), time.Now().UTC())
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "unable to share food")
		return
	}
	utility.Respond(w, http.StatusOK, "OK")
}

func deleteFood(w http.ResponseWriter, r *http.Request) {
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}
}

func updateFood(w http.ResponseWriter, r *http.Request) {
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}
}
