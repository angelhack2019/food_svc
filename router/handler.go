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
	// curl http://localhost:8181/foods/12345
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
	// curl http://localhost:8181/foods?tags=fruits,apple
	// curl http://localhost:8181/foods
	// curl http://localhost:8181/foods?user_uuid=a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, errStrUnableToReachDB)
		return
	}
	tags := r.FormValue("tags")
	userUUID := r.FormValue("user_uuid")

	if strings.TrimSpace(tags) == "" && strings.TrimSpace(userUUID) == "" {
		command := `
				SELECT uuid, pic_url, exp_date, created_date, user_uuid 
				FROM dough_you.foods
				`
		row, err := postgresDB.Query(command)
		if err != nil {
			utility.RespondWithError(w, http.StatusInternalServerError, "unable to get all foods with tag")
			return
		}
		foods := []map[string]string{}
		for row.Next() {
			var foodUUID, picUrl, userUUID string
			var expDate, createdDate time.Time

			err := row.Scan(&foodUUID, &picUrl, &expDate, &createdDate, &userUUID)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan food with tag")
				return
			}
			command := `
				SELECT name
				FROM dough_you.tags
				WHERE dough_you.tags.food_uuid  = $1
				`
			j, err := postgresDB.Query(command, foodUUID)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to get all foods with tag")
				return
			}

			names := []string{}
			for j.Next() {
				var name string
				err := j.Scan(&name)
				if err != nil {
					utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan food with tag")
					return
				}
				names = append(names, name)
			}

			food := map[string]string{
				"uuid":         foodUUID,
				"pic_url":      picUrl,
				"exp_date":     expDate.String(),
				"created_date": createdDate.String(),
				"user_uuid":    userUUID,
				"tags":         strings.Join(names, " "),
			}
			foods = append(foods, food)
		}
		utility.RespondWithJSON(w, http.StatusOK, foods)
		return
	} else if strings.TrimSpace(userUUID) != "" {
		command := `
				SELECT uuid, pic_url, exp_date, created_date, user_uuid 
				FROM dough_you.foods
				WHERE user_uuid = $1
				`
		row, err := postgresDB.Query(command, userUUID)
		if err != nil {
			utility.RespondWithError(w, http.StatusInternalServerError, "unable to get all foods with tag")
			return
		}
		foods := []map[string]string{}
		for row.Next() {
			var foodUUID, picUrl, userUUID string
			var expDate, createdDate time.Time

			err := row.Scan(&foodUUID, &picUrl, &expDate, &createdDate, &userUUID)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan user food with tag")
				return
			}
			command := `
				SELECT name
				FROM dough_you.tags
				WHERE dough_you.tags.food_uuid  = $1
				`
			j, err := postgresDB.Query(command, foodUUID)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to get all foods with tag")
				return
			}

			names := []string{}
			for j.Next() {
				var name string
				err := j.Scan(&name)
				if err != nil {
					utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan food with tag")
					return
				}
				names = append(names, name)
			}

			food := map[string]string{
				"uuid":         foodUUID,
				"pic_url":      picUrl,
				"exp_date":     expDate.String(),
				"created_date": createdDate.String(),
				"user_uuid":    userUUID,
				"tags":         strings.Join(names, " "),
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
				SELECT uuid, pic_url, exp_date, created_date, user_uuid
				FROM dough_you.foods
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
			var foodUUID, picUrl, userUUID string
			var expDate, createdDate time.Time

			err := row.Scan(&foodUUID, &picUrl, &expDate, &createdDate, &userUUID)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan food with tag")
				return
			}
			command := `
				SELECT name
				FROM dough_you.tags
				WHERE dough_you.tags.food_uuid = $1
				`
			j, err := postgresDB.Query(command, foodUUID)
			if err != nil {
				utility.RespondWithError(w, http.StatusInternalServerError, "unable to get all foods with tag")
				return
			}

			names := []string{}
			for j.Next() {
				var name string
				err := j.Scan(&name)
				if err != nil {
					utility.RespondWithError(w, http.StatusInternalServerError, "unable to scan food with tag")
					return
				}
				names = append(names, name)
			}

			food := map[string]string{
				"uuid":         foodUUID,
				"pic_url":      picUrl,
				"exp_date":     expDate.String(),
				"created_date": createdDate.String(),
				"user_uuid":    userUUID,
				"tags":         strings.Join(names, " "),
			}
			foods = append(foods, food)
		}
	}

	utility.RespondWithJSON(w, http.StatusOK, foods)
}

func shareFood(w http.ResponseWriter, r *http.Request) {
	//	curl \
	//	-F "exp_date=1563073799" \
	//	-F "tags=fruits,magic" \
	//	-F "user_uuid=a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11" \
	//	-F "image=@dede.jpeg" \
	//http://localhost:8181/food
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
	userUUID := r.FormValue("user_uuid")
	t, err := strconv.Atoi(expDate)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "wrong expiration date")
		return
	}

	if strings.TrimSpace(tags) == "" {
		utility.RespondWithError(w, http.StatusInternalServerError, "empty tags")
		return
	}

	command := `
				INSERT INTO dough_you.foods(
					uuid, pic_url, exp_date, created_date, user_uuid
				) VALUES($1, $2, $3, $4, $5)
				`
	_, err = postgresDB.Exec(command, foodUUID, link, time.Unix(int64(t), 0), time.Now().UTC(), userUUID)
	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "unable to insert food")
		return
	}

	tagsSlice := strings.Split(tags, ",")
	if len(rekogTags) != 0 {
		tagsSlice = append(tagsSlice, rekogTags...)
	}
	for i := 0; i < len(tagsSlice); i += 1 {
		tagsSlice[i] = strings.ToLower(tagsSlice[i])
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
