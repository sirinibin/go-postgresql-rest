package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"gitlab.com/sirinibin/go-postgresql-rest/migrations"
)

const (
	username = "root"
	password = "123"
	hostname = "localhost"
	port     = 5432
	dbname   = "golang_rest"
)

var DB *sql.DB

func dsn(dbName string) string {
	//return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbName)
	return fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		hostname, port, username, password, dbname)
}

func dbConnection() (*sql.DB, error) {
	/*
		db, err := sql.Open("postgres", dsn(""))
		if err != nil {
			log.Printf("Error %s when opening DB\n", err)
			return nil, err
		}
	*/
	//defer db.Close()

	/*
		ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelfunc()
		res, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+dbname)
		if err != nil {
			log.Printf("Error %s when creating DB\n", err)
			return nil, err
		}
		_, err = res.RowsAffected()
		if err != nil {
			log.Printf("Error %s when fetching rows", err)
			return nil, err
		}
	*/
	//log.Printf("rows affected %d\n", no)

	//db.Close()
	db, err := sql.Open("postgres", dsn(dbname))
	if err != nil {
		log.Printf("Error %s when opening DB", err)
		return nil, err
	}
	//defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	err = db.PingContext(ctx)
	if err != nil {
		log.Printf("Errors %s pinging DB", err)
		return nil, err
	}
	log.Printf("Connected to DB %s successfully\n", dbname)
	return db, nil
}

func InitPostgreSQL() {

	/*
		psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	*/

	var err error
	DB, err = dbConnection()
	if err != nil {
		log.Printf("Error %s when getting db connection", err)
		return
	}

	migrations.Run(DB)

	//Creating Authcode
	os.Setenv("ACCESS_SECRET", "jdnfksdmfksd")

}
