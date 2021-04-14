package migrations

import (
	"context"
	"database/sql"
	"log"
	"time"
	//_ "github.com/go-sql-driver/mysql"
)

func CreateUserTable(db *sql.DB) (sql.Result, error) {
	query := `CREATE SEQUENCE IF NOT EXISTS users_seq;

	CREATE TABLE IF NOT EXISTS users (
			id int NOT NULL PRIMARY KEY DEFAULT NEXTVAL ('users_seq'),
			name varchar(200) DEFAULT NULL,
			username varchar(200) DEFAULT NULL,
			email varchar(200) DEFAULT NULL,
			password varchar(200) DEFAULT NULL,
			created_at timestamp(0) DEFAULT NULL,
			updated_at timestamp(0) DEFAULT NULL
		  )  ;`

	/*query := `CREATE TABLE IF NOT EXISTS user (
		id int NOT NULL PRIMARY KEY AUTO_INCREMENT,
		name varchar(200) DEFAULT NULL,
		username varchar(200) DEFAULT NULL,
		email varchar(200) DEFAULT NULL,
		password varchar(200) DEFAULT NULL,
		created_at datetime DEFAULT NULL,
		updated_at datetime DEFAULT NULL
	  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;`
	*/

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()

	res, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating user table", err)

	}
	return res, err
}
