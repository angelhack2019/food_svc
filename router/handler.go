package router

import (
	"database/sql"
	"github.com/angelhack2019/lib/utility"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

func getFood(w http.ResponseWriter, r *http.Request) {
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}

	vars := mux.Vars(r)
	foodUUID := vars["uuid"]
	command := `
				SELECT uuid, pic_url, exp_date, created_date 
				FROM dough_you.foods
				WHERE dough_you.foods.uuid = $1
				`
	row, err := postgresDB.Query(command, foodUUID)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "unable to get a food with tag")
		return
	}
	for row.Next() {
		var uuid, picUrl string
		var expDate, createdDate time.Time

		err := row.Scan(&uuid, &picUrl, &expDate, &createdDate)
		if err != nil {
			utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan a food with tag")
			return
		}
		food := map[string]string{
			"uuid":         uuid,
			"pic_url":      picUrl,
			"exp_date":     expDate.String(),
			"created_date": createdDate.String(),
		}
		utility.RespondWithJSON(w, http.StatusOK, food)
	}
	return
}

func getFoods(w http.ResponseWriter, r *http.Request) {
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}
	tags := r.FormValue("tags")

	if strings.TrimSpace(tags) == "" {
		command := `
				SELECT uuid, pic_url, exp_date, created_date 
				FROM dough_you.foods
				`
		row, err := postgresDB.Query(command)
		if err != nil {
			utility.RespondWithError(w, http.StatusInternalServerError, "unable to get all foods with tag")
			return
		}
		foods := []map[string]string{}
		for row.Next() {
			var uuid, picUrl string
			var expDate, createdDate time.Time

			err := row.Scan(&uuid, &picUrl, &expDate, &createdDate)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan food with tag")
				return
			}
			food := map[string]string{
				"uuid":         uuid,
				"pic_url":      picUrl,
				"exp_date":     expDate.String(),
				"created_date": createdDate.String(),
			}
			foods = append(foods, food)
		}
		utility.RespondWithJSON(w, http.StatusOK, foods)
		return
	}

	rows := []*sql.Rows{}
	tagsSlice := strings.Split(tags, ",")
	for _, tag := range tagsSlice {
		command := `
				SELECT uuid, pic_url, exp_date, created_date FROM dough_you.foods
				INNER JOIN dough_you.tags
				ON dough_you.foods.uuid = dough_you.tags.food_uuid
				WHERE dough_you.tags.name = $1
				`
		row, err := postgresDB.Query(command, tag)
		if err != nil {
			utility.RespondWithError(w, http.StatusInternalServerError, "unable to get food with tag")
			return
		}
		rows = append(rows, row)
	}

	foods := []map[string]string{}
	for _, row := range rows {
		for row.Next() {
			var uuid, picUrl string
			var expDate, createdDate time.Time

			err := row.Scan(&uuid, &picUrl, &expDate, &createdDate)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan food with tag")
				return
			}
			food := map[string]string{
				"uuid":         uuid,
				"pic_url":      picUrl,
				"exp_date":     expDate.String(),
				"created_date": createdDate.String(),
			}
			foods = append(foods, food)
		}
	}

	utility.RespondWithJSON(w, http.StatusOK, foods)
}

func shareFood(w http.ResponseWriter, r *http.Request) {
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}
	picURL := r.FormValue("pic_url")
	expDate := r.FormValue("exp_date")
	tags := r.FormValue("tags")

	t, err := strconv.Atoi(expDate)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "wrong expiration date")
		return
	}
	// TODO post to AWS S3

	if strings.TrimSpace(tags) == "" {
		utility.RespondWithError(w, http.StatusInternalServerError, "empty tags")
		return
	}

	foodUUID := uuid.New().String()
	command := `
				INSERT INTO dough_you.foods(
					uuid, pic_url, exp_date, created_date 
				) VALUES($1, $2, $3, $4)
				`
	_, err = postgresDB.Exec(command, foodUUID, picURL, time.Unix(int64(t), 0), time.Now().UTC())
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "unable to insert food")
		return
	}

	tagsSlice := strings.Split(tags, ",")
	for _, tag := range tagsSlice {
		command := `
				INSERT INTO dough_you.tags(
					food_uuid, name
				) VALUES($1, $2)
				`
		_, err = postgresDB.Exec(command, foodUUID, tag)
		if err != nil {
			utility.RespondWithError(w, http.StatusInternalServerError, "unable to insert tags")
			return
		}
	}

	utility.Respond(w, http.StatusOK, "OK")
}
