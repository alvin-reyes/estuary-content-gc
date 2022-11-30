package content_gc

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

type ContentGc struct {
	BaseGC BaseGC
	DryRun *bool
}

type ToCleanUp struct {
	Host string
	ID   string
}

// A method of the ContentGc struct.
func (c ContentGc) Run() {

	var countNumberOfMarkedForDeletion int64
	var countNumberOfValidContent int64
	results := &[]ToCleanUp{}
	currentTime := time.Now()
	last3Month := currentTime.AddDate(0, -3, 0)
	timeLayout := "2006-01-02"
	stringDate := last3Month.Format(timeLayout)

	fmt.Println("Running cleanup for content marked for deletion before: ", stringDate)

	c.BaseGC.DB.Raw("select s.host, c.id from contents as c, shuttles as s where c.location = s.handle and (c.created_at between '2000-01-01' and ?) and c.pinning group by s.host, c.id", stringDate).Scan(results)

	fmt.Println("Running through ", len(*results), " results")
	for _, result := range *results {

		if result.Host == "shuttle-3.estuary.tech" { // shuttle-3 don't exist anymore. don't even try
			fmt.Println("Record: ", result.Host, result.ID, "SHUTTLE-3_DOES_NOT_EXIST_ANYMORE")
			c.markItAsFail(result.ID)
			continue
		}

		client := &http.Client{}
		req, _ := http.NewRequest("GET", "https://"+result.Host+ShuttleCheckEndpoint+result.ID, nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+viper.Get("API_KEY").(string))
		res, err := client.Do(req)

		if err != nil { // skip if error
			fmt.Println("Record: ", result.Host, result.ID, err, "ERROR_ON_SHUTTLE_REQUEST_CHECK")
			continue
		}
		if res.StatusCode != http.StatusOK { // mark it!
			fmt.Println("Record: ", result.Host, result.ID, res.StatusCode, "MARK_AS_FAILED")
			countNumberOfMarkedForDeletion++
			c.markItAsFail(result.ID)
		} else {
			fmt.Println("Record: ", result.Host, result.ID, res.StatusCode, "GOOD")
			countNumberOfValidContent++
		}
	}

	fmt.Println("Number of marked for deletion: ", countNumberOfMarkedForDeletion)
	fmt.Println("Number of valid content: ", countNumberOfValidContent)
}

// Marking the content as failed.

func (c ContentGc) markItAsFail(contentId string) {
	if !*c.DryRun {
		c.BaseGC.DB.Raw("update contents set pinning = false, active=false, failed = true where id = ?", contentId)
	}
}
