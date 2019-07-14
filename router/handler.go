package router

import (
	"database/sql"
	"fmt"
	"github.com/angelhack2019/lib/utility"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	errStrUnableToReachDB = "unable to connect to db"
)

var (
	pg               string
	connectionString string
	postgresDB       *sql.DB
)

func init() {
	viper.BindEnv("PG")
	pg = viper.GetString("PG")
	connectionString = fmt.Sprintf(
		"host=%s user=default password=default dbname=dough_you sslmode=disable port=5432",
		pg,
	)

}
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
	// curl http://localhost:8080/foods/12345
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
	// curl http://localhost:8080/foods?tags=fruits,apple
	// curl http://localhost:8080/foods
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
	// curl \
	//-F "exp_date=1563073799" \
	//-F "tags=fruits,magic" \
	//-F "image=@image.jpg" \
	//http://localhost:8080/food
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}

	foodUUID := uuid.New().String()

	link, rekogTags, err := uploadFile(w, r, foodUUID)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

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

	command := `
				INSERT INTO dough_you.foods(
					uuid, pic_url, exp_date, created_date 
				) VALUES($1, $2, $3, $4)
				`
	_, err = postgresDB.Exec(command, foodUUID, link, time.Unix(int64(t), 0), time.Now().UTC())
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "unable to insert food")
		return
	}

	tagsSlice := strings.Split(tags, ",")
	if len(rekogTags) != 0 {
		tagsSlice = append(tagsSlice, rekogTags...)
	}
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
