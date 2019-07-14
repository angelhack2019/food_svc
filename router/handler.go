package router

import (
	"database/sql"
	"fmt"
	"github.com/angelhack2019/lib/utility"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

var (
	connectionString string
	postgresDB       *sql.DB
)

func init() {
	connectionString = "host=192.168.0.230 user=default password=default dbname=dough_you sslmode=disable port=5432"
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
	//http://localhost:8080/foods?tags=apple%2Cfruits&name=john
	fmt.Fprintln(w, "Get Foods")
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Unable to connect to db")
	}
	tags := r.URL.Query().Get("tags")
	fmt.Fprintln(w, tags)
	name := r.URL.Query().Get("name")
	fmt.Fprintln(w, name)
}

func shareFood(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Share Food")
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Unable to connect to db")
	}

	expDate := r.URL.Query().Get("exp_date")
	picURL := r.URL.Query().Get("pic_url")

	command := `
				INSERT INTO food_svc.dough_you.foods(
					uuid, pic_url, exp_date, created_date 
				) VALUES($1, $2, $3, $4)
				`

	_, err := postgresDB.Exec(command, uuid.New().String(), picURL, expDate, time.Now().UTC())

	if err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Unable to share food")
	}

	utility.Respond(w, http.StatusOK, "OK")
}

func deleteFood(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Delete Food")
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Unable to connect to db")
	}
}

func updateFood(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Update food")
	if err := refreshDBConnection(); err != nil {
		utility.RespondWithError(w, http.StatusInternalServerError, "Unable to connect to db")
	}
}
