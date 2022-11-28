package main

import (
	"flag"
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
	DB                   *gorm.DB
	DryRun               *bool
	ShuttleCheckEndpoint = "/content/read/" // returns 404 if not available, 200 if content is available.
)

func main() {

	DryRun = flag.Bool("dryrun", true, "dry run to check stats before cleaning up")
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

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	runCleanup()

}

func runCleanup() {

	var countNumberOfMarkedForDeletion int64
	var countNumberOfValidContent int64
	results := &[]ToCleanUp{}
	currentTime := time.Now()
	last3Month := currentTime.AddDate(0, -3, 0)
	timeLayout := "2006-01-02"
	stringDate := last3Month.Format(timeLayout)

	fmt.Println("Running cleanup for content marked for deletion before: ", stringDate)

	DB.Raw("select s.host, c.id from contents as c, shuttles as s where c.location = s.handle and (c.created_at between '2000-01-01' and ?) and c.pinning group by s.host, c.id", stringDate).Scan(results)

	fmt.Println("Running through ", len(*results), " results")
	for _, result := range *results {

		if result.Host == "shuttle-3.estuary.tech" { // shuttle-3 don't exist anymore. don't even try
			fmt.Println("Record: ", result.Host, result.ID, "SHUTTLE-3_DOES_NOT_EXIST_ANYMORE")
			markIt(result.ID)
			continue
		}

		client := &http.Client{}
		req, _ := http.NewRequest("GET", "https://"+result.Host+ShuttleCheckEndpoint+result.ID, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+viper.Get("API_KEY").(string))
		res, err := client.Do(req)

		if err != nil { // skip if error
			fmt.Println("Record: ", result.Host, result.ID, err, "ERROR_ON_SHUTTLE_REQUES_CHECK")
			continue
		}
		if res.StatusCode != http.StatusOK { // mark it!
			fmt.Println("Record: ", result.Host, result.ID, res.StatusCode, "MARK_AS_FAILED")
			countNumberOfMarkedForDeletion++
			markIt(result.ID)
		} else {
			fmt.Println("Record: ", result.Host, result.ID, res.StatusCode, "GOOD")
			countNumberOfValidContent++
		}
	}

	fmt.Println("Number of marked for deletion: ", countNumberOfMarkedForDeletion)
	fmt.Println("Number of valid content: ", countNumberOfValidContent)
}

func markIt(contentId string) {
	if !*DryRun {
		DB.Raw("update contents set pinning = false, active=false, failed = true where id = ?", contentId)
	}
}
