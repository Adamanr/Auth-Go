package main

import (
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/buntdb"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Registration_User struct {
	Login          string
	Password       string
	SecondPassword string `json:"secondPassword"`
}

var db, _ = buntdb.Open(":memory:")

func main() {

	router := gin.New()
	router.POST("/auth", authorization)
	router.POST("/registration", registration)
	router.POST("/show_users", show_users)
	dbInit()
	log.Fatal(router.Run(":3031"))
}

func show_users(context *gin.Context) {
	db.View(func(tx *buntdb.Tx) error {
		tx.Ascend("logins", func(key, value string) bool {
			fmt.Printf("%s: %s\n", key, value)
			return true
		})
		return nil
	})
}

type LoginResponse struct {
	Token string `json:"token"`
}

func authorization(c *gin.Context) {
	user := User{}
	err := c.BindJSON(&user)
	if err != nil {
		log.Println(err)
		return
	}
	var user_id int
	var user_password string
	db.View(func(tx *buntdb.Tx) error {
		tx.AscendEqual("logins", user.Login, func(key, value string) bool {
			user_id, _ = strconv.Atoi(strings.TrimRight(strings.TrimLeft(key, "user:"), ":login"))
			return true
		})
		tx.AscendKeys(fmt.Sprintf("user:%d:password", user_id), func(key, value string) bool {
			user_password = value
			return true
		})
		return nil
	})
	var sx buntdb.Tx
	if user.Password == user_password {
		fmt.Println("Пароли совпадают")
		token := GenerateSecureToken(16)
		var rl = LoginResponse{Token: token}
		fmt.Println(token)
		sx.Set(fmt.Sprintf("login:%d:token", user_id), token, nil)
		c.JSON(200, rl)
	} else {
		fmt.Println("Пароли не совпадают")
	}
}
func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
func registration(c *gin.Context) {
	user := Registration_User{}
	err := c.BindJSON(&user)
	if err != nil {
		log.Println(err)
	}
	if user.Password == user.SecondPassword {
		log.Println("Пароли одинаковы")
		dbAdd(user.Login, user.Password)
	} else {
		log.Println("Пароли не подходят")
	}
}

func dbInit() {
	db.CreateIndex("logins", "user:*:login", buntdb.IndexString)
	db.CreateIndex("passwords", "user:*:password", buntdb.IndexString)
	db.Update(func(tx *buntdb.Tx) error {
		tx.Set("user:1:login", "Egor", nil)
		tx.Set("user:1:password", "456", nil)
		tx.Set("user:2:login", "Artem", nil)
		tx.Set("user:2:password", "123", nil)
		tx.Set("user:3:login", "Dima", nil)
		tx.Set("user:3:password", "321", nil)
		return nil
	})
}

func dbAdd(login, password string) {
	var users []string
	db.View(func(tx *buntdb.Tx) error {
		tx.DescendKeys("*", func(key, value string) bool {
			users = append(users, key)
			return true
		})
		return nil
	})
	user_id, _ := strconv.Atoi(strings.TrimRight(strings.TrimLeft(users[0], "user:"), ":password"))
	user_id += 1
	db.Update(func(tx *buntdb.Tx) error {
		tx.Set(fmt.Sprintf("user:%d:login", user_id), login, nil)
		tx.Set(fmt.Sprintf("user:%d:password", user_id), password, nil)
		return nil
	})
}

func dbGet(user_name string) {
	db.View(func(tx *buntdb.Tx) error {
		var user_id string
		fmt.Println("Order by last name")
		tx.AscendEqual("logins", user_name, func(key, value string) bool {
			user_id = strings.TrimRight(strings.TrimLeft(key, "user:"), ":login")
			fmt.Printf("%s \n")
			return true
		})

		return nil
	})
}
