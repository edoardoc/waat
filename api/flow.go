package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var (
	signKey = []byte("JKJKKJKLtestingwaat")
	db      *bun.DB
)

type Apiuser struct {
	Email     string `bun:",pk"`
	Password  string `json:"password,omitempty"`
	Firstname string
	Lastname  string
}
type ApiUserClaims struct {
	*jwt.StandardClaims
	Apiuser
}

// createSchema creates database schema for User and Story models.
func createSchema(db *bun.DB) error {
	ctx := context.Background()

	models := []interface{}{
		(*Apiuser)(nil),
	}

	for _, model := range models {
		log.Printf("creating table %v ", reflect.TypeOf(model))

		_, err := db.NewCreateTable().Model(model).Exec(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func createToken(email string) (time.Time, string, error) {
	expirationTime := time.Now().Add(30 * time.Minute)
	t := jwt.New(jwt.SigningMethodHS256)

	t.Claims = &ApiUserClaims{
		&jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
		Apiuser{Email: email},
	}
	newtoken, err := t.SignedString(signKey)
	return expirationTime, newtoken, err
}

func decodeToken(tokenString string) (string, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &ApiUserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// since we only use the one private key to sign the tokens,
		// we also only use its public counter part to verify
		return signKey, nil
	})
	claims := token.Claims.(*ApiUserClaims)
	return claims.Apiuser.Email, err
}

func saveApiUser(item Apiuser) error {
	ctx := context.Background()
	_, err := db.NewInsert().
		Model(&item).
		On("CONFLICT (email) DO UPDATE").
		Set("Firstname = EXCLUDED.Firstname").
		Set("Lastname = EXCLUDED.Lastname").
		Exec(ctx)
	return err
}

func updateApiUser(item Apiuser) error {
	ctx := context.Background()
	_, err := db.NewUpdate().
		Model(&item).
		Set("firstname = ?", item.Firstname).
		Set("lastname = ?", item.Lastname).
		WherePK().
		// Where("email = ?", item.Email).
		Exec(ctx)
	return err
}

func checkLoginApiUser(item Apiuser) (bool, error) {
	ctx := context.Background()
	count, err := db.NewSelect().
		Model(&item).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("email = ?", item.Email).Where("password = ?", item.Password)
		}).
		Count(ctx)
	return count == 1, err
}

func receivedUser(r *http.Request) (Apiuser, error) {
	decoder := json.NewDecoder(r.Body)
	var oneUser Apiuser
	err := decoder.Decode(&oneUser)
	return oneUser, err
}

// curl -v --header "Content-Type: application/json" --request POST --data '{"email":"ridleys@gmail.com","password":"233223edfsdf","firstname":"Edoardo","lastname":"Ceccarelli"}' http://localhost:8080/api/signup
func Signup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		log.Printf("serving /Signup ")

		oneUser, err := receivedUser(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			// a valid user was received (fair assumption, more checks are needed)

			// ** TOKEN GENERATION
			_, tokenString, err := createToken(oneUser.Email)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				err = saveApiUser(oneUser)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("x-auth-token", tokenString)
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// curl -v --header "Content-Type: application/json" --request POST --data '{"email":"ridleys@gmail.com","password":"test123"}' http://localhost:8080/api/login
func Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		log.Printf("serving /Login ")

		oneUser, err := receivedUser(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			credentialsOk, err := checkLoginApiUser(oneUser)
			if credentialsOk && err == nil {
				// ** TOKEN GENERATION
				_, tokenString, err := createToken(oneUser.Email)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Header().Set("x-auth-token", tokenString)
					w.WriteHeader(http.StatusOK)
				}
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// curl -v --header "X-Auth-Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDYwNjU2NDgsIkVtYWlsIjoicmlkbGV5c0BnbWFpbC5jb20iLCJQYXNzd29yZCI6IiIsIkZpcnN0bmFtZSI6IiIsIkxhc3RuYW1lIjoiIn0.wbJl8b1xjsTavzk8g4mumDOt3NROHXv8Z-AoCBG1tvM" --header "Content-Type: application/json" --request PUT --data '{"firstname":"wwwEdoardo","lastname":"Ceccadddrelli"}' http://localhost:8080/api/users
// curl -v --header "X-Auth-Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDYwNjU2NDgsIkVtYWlsIjoicmlkbGV5c0BnbWFpbC5jb20iLCJQYXNzd29yZCI6IiIsIkZpcnN0bmFtZSI6IiIsIkxhc3RuYW1lIjoiIn0.wbJl8b1xjsTavzk8g4mumDOt3NROHXv8Z-AoCBG1tvM" --header "Content-Type: application/json" --request GET http://localhost:8080/api/users
func Users(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		log.Printf("serving Get /Users ")
		receivedToken := r.Header.Get("x-auth-token")
		_, err := decodeToken(receivedToken) // ** a valid token means ok to give the list
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			// ** Apiuser LIST USERS
			ctx := context.Background()
			var listUser []*Apiuser
			err = db.NewSelect().
				Model((*Apiuser)(nil)).
				ColumnExpr("email").
				ColumnExpr("firstname").
				ColumnExpr("lastname").
				Scan(ctx, &listUser)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				formattedOut, err := json.MarshalIndent(listUser, "", "    ")
				// formattati, err := json.Marshal(listUser)
				if err != nil {
					fmt.Println(err)
					return
				}
				w.Write(formattedOut)
				w.WriteHeader(http.StatusOK)
			}
		}
	case http.MethodPut:
		log.Printf("serving Put /Users ")
		receivedToken := r.Header.Get("x-auth-token")
		log.Printf("receivedToken : %v\n", receivedToken)
		whichUser, err := decodeToken(receivedToken) // ** DECODING TOKEN
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			oneUser, err := receivedUser(r)
			oneUser.Email = whichUser // to make sure the seek happens on the token value
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				err = updateApiUser(oneUser)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			}
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	// ** DB CONNECTION
	dsn := "postgres://waatusr:@database:5432/waat?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db = bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	if len(os.Args) >= 2 {
		if os.Args[1] == "createDb" {
			fmt.Println("creating database...")
			// ** DB CREATION
			err := createSchema(db)
			if err != nil {
				log.Fatal(err)
			}
		} else if os.Args[1] == "dropDb" {
			fmt.Println("N/D")
		} else {
			fmt.Printf("unrecognised parameter: %v", os.Args[1])
		}
		return
	}

	log.Printf("starting api ")
	http.HandleFunc("/api/signup", Signup)
	http.HandleFunc("/api/login", Login)
	http.HandleFunc("/api/users", Users)
	http.ListenAndServe(":8080", nil)
}
