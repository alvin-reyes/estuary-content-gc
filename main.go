package main

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type ToCleanUp struct {
	Host string
	ID   string
}

var (
	ShuttleCheckEndpoint = "/content/read/" // returns 404 if not available, 200 if content is available.
)

func main() {
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()

	dbHost, okHost := viper.Get("DB_HOST").(string)
	dbUser, okUser := viper.Get("DB_USER").(string)
	dbPass, okPass := viper.Get("DB_PASS").(string)
	dbName, okName := viper.Get("DB_NAME").(string)
	dbPort, okPort := viper.Get("DB_PORT").(string)
	if !okHost || !okUser || !okPass || !okName || !okPort {
		panic("invalid database configuration")
	}

	dsn := "host=" + dbHost + " user=" + dbUser + " password=" + dbPass + " dbname=" + dbName + " port=" + dbPort + " sslmode=disable TimeZone=Asia/Shanghai"

	DB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	results := &[]ToCleanUp{}
	currentTime := time.Now()
	last3Month := currentTime.AddDate(0, -3, 0)
	timeLayout := "2006-01-02"
	stringDate := last3Month.Format(timeLayout)

	DB.Raw("select s.host, c.id from contents as c, shuttles as s where c.location = s.handle and (c.created_at between '2000-01-01' and ?) and c.pinning group by s.host, c.id", stringDate).Scan(results)

	for _, result := range *results {
		fmt.Println(result.ID + " " + result.Host)

		response, err := http.Get("https://" + result.Host + ShuttleCheckEndpoint + result.ID)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(response.StatusCode)
		if response.StatusCode != http.StatusOK { // mark it!
			fmt.Println("Marking " + result.ID + " " + result.Host)
			//Commented out for now: DB.Raw("update contents set pinning = false where id = ?", result.ID)
		}

		fmt.Println(response.Body)
	}
}
