package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	guuid "github.com/google/uuid"
	"github.com/gorilla/handlers"
	_ "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

const (
	host     = "18.140.67.82"
	port     = 8976
	user     = "test"
	password = "kriyatest123"
	dbname   = "kriya_test"
)

type datauser struct {
	JSONUser json.RawMessage `json:"data"`
	RoleName string          `json:"role_name"`
	Creator  string          `json:"username"`
	Usercode string          `json:"usercode"`
}

type jsonuser struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
	Username  string `json:"username"`
	IsActive  bool   `json:"is_active"`
	LastName  string `json:"last_name"`
	UserCode  string `json:"user_code"`
	FirstName string `json:"first_name"`
	LastLogin int    `json:"last_login"`
}

type listuser struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleName string `json:"role_name"`
}

type responseerror struct {
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Path      string    `json:"path"`
}

type responsesuccess struct {
	Timestamp time.Time `json:"timestamp"`
	Status    int       `json:"status"`
	Success   string    `json:"Success"`
	Message   string    `json:"message"`
	Path      string    `json:"path"`
}

func main() {
	router := mux.NewRouter()
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Content-Length", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	origins := handlers.AllowedOrigins([]string{"*"})
	// HEALTH CHECK
	router.HandleFunc("", defaultHandler)
	router.HandleFunc("/", defaultHandler)
	r := router.PathPrefix("/back-end").Subrouter()
	// EXTENDED HEALTH CHECK
	r.HandleFunc("", defaultHandler)
	r.HandleFunc("/", defaultHandler)
	//Make an api to update user (only admin can access)
	r.HandleFunc("/createUser", createUser).Methods("POST")
	//Make an api to update user (only admin can access)
	r.HandleFunc("/updateUser", updateUser).Methods("POST")
	//Make an api to delete user (only admin can access)
	r.HandleFunc("/deleteUser", deleteUser).Methods("POST")
	//Make an api to get user data with the following structure
	r.HandleFunc("/getUser", getUser).Methods("GET")

	http.ListenAndServe(":8000", handlers.CORS(headers, methods, origins)(router))
}

func connect() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return db, nil
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	db, err := connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	w.Write([]byte("REST API"))
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	var datauser datauser
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&datauser)
	dt := time.Now()
	usercode := datauser.Usercode
	creator := datauser.Creator
	// jsonuser := datauser.JSONUser

	db, err := connect()
	if err != nil {
		fmt.Println(err.Error())
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Gagal, Terjadi kesalahan saat Koneksi Database",
			Path:      "delete_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}
	defer db.Close()

	var count int

	errs := db.QueryRow("select count(*) from users a ,roles b where a.role_id = b.id and b.data ->> 'role_name' = 'Admin' and a.data ->> 'username' = $1", creator).Scan(&count)
	switch {
	case errs != nil:
		// log.Println("error disini 1")
		// log.Fatal(errs)
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Gagal, Terjadi kesalahan saat Delete User",
			Path:      "delete_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}

	if count == 1 {

		_, err := db.Query("UPDATE users SET deleted_at = now() where data ->> 'user_code' = $1 ", usercode)
		if err != nil {
			log.Println("error insert")
			fmt.Println(err.Error())
			u := responseerror{
				Timestamp: dt,
				Status:    201,
				Error:     "Gagal",
				Message:   "Gagal, Terjadi kesalahan saat Delete User",
				Path:      "delete_user",
			}
			w.WriteHeader(http.StatusOK)
			prettyJSON, err := json.MarshalIndent(u, "", "    ")
			if err != nil {
				log.Fatal("Failed to generate json", err)
			}
			fmt.Fprintf(w, string(prettyJSON))
		} else {
			u := responsesuccess{
				Timestamp: dt,
				Status:    200,
				Success:   "Berhasil",
				Message:   "Berhasil Delete User",
				Path:      "delete_user",
			}
			w.WriteHeader(http.StatusOK)
			prettyJSON, err := json.MarshalIndent(u, "", "    ")
			if err != nil {
				log.Fatal("Failed to generate json", err)
			}
			fmt.Fprintf(w, string(prettyJSON))
		}

	} else {
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Anda Tidak Memiliki Access Admin",
			Path:      "delete_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}

}

func updateUser(w http.ResponseWriter, r *http.Request) {
	var datauser datauser
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&datauser)
	dt := time.Now()
	usercode := datauser.Usercode
	creator := datauser.Creator
	jsonuser := datauser.JSONUser

	db, err := connect()
	if err != nil {
		fmt.Println(err.Error())
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Gagal, Terjadi kesalahan saat Koneksi Database",
			Path:      "update_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}
	defer db.Close()

	var count int

	errs := db.QueryRow("select count(*) from users a ,roles b where a.role_id = b.id and b.data ->> 'role_name' = 'Admin' and a.data ->> 'username' = $1", creator).Scan(&count)
	switch {
	case errs != nil:
		// log.Println("error disini 1")
		// log.Fatal(errs)
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Gagal, Terjadi kesalahan saat Update User",
			Path:      "update_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}

	if count == 1 {

		_, err := db.Query("UPDATE users SET data = $1, updated_at = now() where data ->> 'user_code' = $2 ", jsonuser, usercode)
		if err != nil {
			// log.Println("error insert")
			// fmt.Println(err.Error())
			u := responseerror{
				Timestamp: dt,
				Status:    201,
				Error:     "Gagal",
				Message:   "Gagal, Terjadi kesalahan saat Update User",
				Path:      "update_user",
			}
			w.WriteHeader(http.StatusOK)
			prettyJSON, err := json.MarshalIndent(u, "", "    ")
			if err != nil {
				log.Fatal("Failed to generate json", err)
			}
			fmt.Fprintf(w, string(prettyJSON))
		} else {
			u := responsesuccess{
				Timestamp: dt,
				Status:    200,
				Success:   "Berhasil",
				Message:   "Berhasil Update User",
				Path:      "update_user",
			}
			w.WriteHeader(http.StatusOK)
			prettyJSON, err := json.MarshalIndent(u, "", "    ")
			if err != nil {
				log.Fatal("Failed to generate json", err)
			}
			fmt.Fprintf(w, string(prettyJSON))
		}

	} else {
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Anda Tidak Memiliki Access Admin",
			Path:      "update_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}

}

func createUser(w http.ResponseWriter, r *http.Request) {
	var datauser datauser
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&datauser)
	dt := time.Now()
	rolename := datauser.RoleName
	creator := datauser.Creator
	jsonuser := datauser.JSONUser

	db, err := connect()
	if err != nil {
		fmt.Println(err.Error())
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Gagal, Terjadi kesalahan saat Koneksi Database",
			Path:      "create_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}
	defer db.Close()

	var count int

	errs := db.QueryRow("select count(*) from users a ,roles b where a.role_id = b.id and b.data ->> 'role_name' = 'Admin' and a.data ->> 'username' = $1", creator).Scan(&count)
	switch {
	case errs != nil:
		// log.Println("error disini 1")
		// log.Fatal(errs)
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Gagal, Terjadi kesalahan saat Create User",
			Path:      "create_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}

	if count == 1 {
		rows, err := db.Query("select distinct id from roles where data ->> 'role_name' = $1 limit 1;", rolename)
		if err != nil {
			// log.Println("error disini 2")
			// fmt.Println(err.Error())
			u := responseerror{
				Timestamp: dt,
				Status:    201,
				Error:     "Gagal",
				Message:   "Gagal, Terjadi kesalahan saat Create User",
				Path:      "create_user",
			}
			w.WriteHeader(http.StatusOK)
			prettyJSON, err := json.MarshalIndent(u, "", "    ")
			if err != nil {
				log.Fatal("Failed to generate json", err)
			}
			fmt.Fprintf(w, string(prettyJSON))
		}
		defer rows.Close()

		for rows.Next() {
			var roleid string
			err = rows.Scan(&roleid)
			if err != nil {
				panic(err.Error())
			}

			id := guuid.New()
			_, err := db.Query("INSERT INTO users (id, data, role_id, created_at,updated_at, deleted_at) values ( $1, $2, $3, now(), now(), null)", id, jsonuser, roleid)
			if err != nil {
				// log.Println("error insert")
				// fmt.Println(err.Error())
				u := responseerror{
					Timestamp: dt,
					Status:    201,
					Error:     "Gagal",
					Message:   "Gagal, Terjadi kesalahan saat Create User",
					Path:      "create_user",
				}
				w.WriteHeader(http.StatusOK)
				prettyJSON, err := json.MarshalIndent(u, "", "    ")
				if err != nil {
					log.Fatal("Failed to generate json", err)
				}
				fmt.Fprintf(w, string(prettyJSON))
			} else {
				u := responsesuccess{
					Timestamp: dt,
					Status:    200,
					Success:   "Berhasil",
					Message:   "Berhasil Create User",
					Path:      "create_user",
				}
				w.WriteHeader(http.StatusOK)
				prettyJSON, err := json.MarshalIndent(u, "", "    ")
				if err != nil {
					log.Fatal("Failed to generate json", err)
				}
				fmt.Fprintf(w, string(prettyJSON))
			}
		}
	} else {
		u := responseerror{
			Timestamp: dt,
			Status:    201,
			Error:     "Gagal",
			Message:   "Anda Tidak Memiliki Access Admin",
			Path:      "create_user",
		}
		w.WriteHeader(http.StatusOK)
		prettyJSON, err := json.MarshalIndent(u, "", "    ")
		if err != nil {
			log.Fatal("Failed to generate json", err)
		}
		fmt.Fprintf(w, string(prettyJSON))
	}

}

func getUser(w http.ResponseWriter, r *http.Request) {
	db, err := connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	rows, err := db.Query("select a.data ->> 'user_code', a.data ->> 'username', a.data ->> 'email', b.data ->> 'role_name' from users a, roles b where a.role_id = b.id and a.deleted_at is null;")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rows.Close()

	data := listuser{}
	arrdata := []listuser{}

	for rows.Next() {

		var usercode, username, email, rolename string
		err = rows.Scan(&usercode, &username, &email, &rolename)
		if err != nil {
			panic(err.Error())
		}

		data.UserID = usercode
		data.Username = username
		data.Email = email
		data.RoleName = rolename
		arrdata = append(arrdata, data)
	}
	jsonData, err := json.Marshal(arrdata)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Fprintf(w, string(jsonData))
}
