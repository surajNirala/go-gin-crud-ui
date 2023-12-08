package Models

import "database/sql"

var db *sql.DB

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
