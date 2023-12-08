package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var baseURL = "http://localhost:8081"

type UserInfo struct {
	ID         int    `form:"id"  db:"id" json:"id"`
	FirstName  string `form:"first_name"  db:"first_name" json:"first_name"`
	LastName   string `form:"last_name"  db:"last_name" json:"last_name"`
	Email      string `form:"email"  db:"email" json:"email"`
	Company    string `form:"company"  db:"company" json:"company"`
	Phone      string `form:"phone"  db:"phone" json:"phone"`
	Url        string `form:"url"  db:"url" json:"url"`
	Shift      string `form:"shift"  db:"shift" json:"shift"`
	Newsletter string `form:"newsletter"  db:"newsletter" json:"newsletter"`
}

func init() {
	var err error
	db, err = sql.Open("mysql", "root@tcp(localhost:3306)/go_crud")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database connected successfully")
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("views/*")
	router.GET("/", indexHandle)
	router.POST("/", storeHandle)
	router.GET("/edit/:id", editHandle)
	router.POST("/edit/:id", updateHandle)
	router.GET("/view/:id", viewHandle)
	router.GET("/delete/:id", deleteHandle)

	router.Run(":" + "8081") // listen and serve on 0.0.0.0:8080
}

func indexHandle(c *gin.Context) {
	users, err := getAllUsers()
	if err != nil {
		c.HTML(500, "form.html", gin.H{"error": err.Error()})
		return
	}
	message := c.Query("message")
	data := gin.H{"Users": users, "Message": message}
	c.HTML(200, "form.html", data)
}

func storeHandle(c *gin.Context) {
	first_name := c.PostForm("first_name")
	last_name := c.PostForm("last_name")
	email := c.PostForm("email")
	company := c.PostForm("company")
	phone := c.PostForm("phone")
	url := c.PostForm("url")
	shift := c.PostForm("shift")
	newsletter := c.PostForm("newsletter")
	insertQuery := "INSERT INTO user_info (first_name, last_name, email, company, phone, url, shift, newsletter) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := db.Exec(insertQuery, first_name, last_name, email, company, phone, url, shift, newsletter)
	if err != nil {
		c.HTML(500, "form.html", gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusMovedPermanently, baseURL+"?message=Record Created Successfully")
	// c.HTML(200, "form.html", data)
}

func getAllUsers() ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT * FROM user_info")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Fetch all rows into a slice
	// var users []UserInfo
	// err = sqlx.StructScan(rows, &users)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Create a slice to hold data dynamically
	var result []map[string]interface{}

	// Create a slice of empty interfaces to hold values
	values := make([]interface{}, len(columns))
	for i := range columns {
		values[i] = new(interface{})
	}

	// Iterate over rows
	for rows.Next() {
		// Scan values into interface slice
		err := rows.Scan(values...)
		if err != nil {
			return nil, err
		}

		// Create a map to store column-value pairs
		record := make(map[string]interface{})

		// Fill the map with column-value pairs
		for i, col := range columns {
			// Convert values to appropriate types
			switch v := (*(values[i].(*interface{}))).(type) {
			case []byte:
				record[col] = string(v)
			default:
				record[col] = v
			}
		}

		// Append the map to the result slice
		result = append(result, record)
	}

	return result, nil
}

func editHandle(c *gin.Context) {
	user_id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid user ID.")
	}
	user, err := getUserByID(user_id)
	if err != nil {
		c.String(http.StatusNoContent, "User Not Found.")
	}
	// Render the edit form HTML page
	c.HTML(http.StatusOK, "edit-form.html", gin.H{"user": user})
}

func viewHandle(c *gin.Context) {
	user_id_param := c.Param("id")
	user_id, err := strconv.Atoi(user_id_param)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid User ID.")
	}
	user, err := getUserByID(user_id)
	if err != nil {
		c.String(http.StatusNoContent, "User Not Found.")
	}
	c.HTML(http.StatusOK, "view-form.html", gin.H{"user": user})
}

func deleteHandle(c *gin.Context) {
	user_id_param := c.Param("id")
	user_id, err := strconv.Atoi(user_id_param)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid User ID.")
	}
	err = deleteUserByID(user_id)
	if err != nil {
		c.String(http.StatusNoContent, "User Not Found.")
	}
	c.Redirect(http.StatusMovedPermanently, baseURL+"?message=Record Delete Successfully")
}

func getUserByID(user_id int) (UserInfo, error) {
	var user UserInfo
	info := db.QueryRow("SELECT * FROM user_info WHERE id = ?", user_id)
	err := info.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Company, &user.Phone, &user.Url, &user.Shift, &user.Newsletter)
	return user, err
}

func updateHandle(c *gin.Context) {
	var form UserInfo
	if err := c.ShouldBind(&form); err != nil {
		c.String(http.StatusBadRequest, "Invalid form data")
		return
	}
	err := updateUser(form)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Error updating user: %s", err.Error()))
		return
	}
	c.Redirect(http.StatusSeeOther, baseURL+"?message= Record Updated Successfully")
}

func updateUser(user UserInfo) error {
	sqlStatement := "UPDATE user_info SET first_name=?, last_name=?, email=?, company=?, phone=?, url=?, shift=?, newsletter=? WHERE id=?"

	// Print the SQL statement and parameters
	fmt.Printf("Executing query: %s\n", sqlStatement)
	fmt.Printf("Parameters: %s, %s, %s, %s, %s, %s, %s, %s, %d\n",
		user.FirstName, user.LastName, user.Email, user.Company, user.Phone, user.Url, user.Shift, user.Newsletter, user.ID)
	// Execute the query
	result, err := db.Exec(sqlStatement,
		user.FirstName, user.LastName, user.Email, user.Company, user.Phone, user.Url, user.Shift, user.Newsletter, user.ID)

	if err != nil {
		// Handle the error (print, log, or return it)
		fmt.Println("Error updating user:", err)
		return err
	}

	numRowsAffected, err := result.RowsAffected()
	if err != nil {
		// Handle the error (print, log, or return it)
		fmt.Println("Error getting rows affected:", err)
		return err
	}

	fmt.Printf("Update successful. Rows affected: %d\n", numRowsAffected)
	return nil
}

func deleteUserByID(user_id int) error {
	_, err := db.Exec("DELETE FROM user_info WHERE id = ?", user_id)
	return err
}
