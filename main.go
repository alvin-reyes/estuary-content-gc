package main

import (
	"content-update-gc/content_gc"
	"flag"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	contentGc := content_gc.ContentGc{
		BaseGC: content_gc.BaseGC{
			DB: DB,
		},
		DryRun: DryRun,
	}
	contentGc.Run() // run it

}
