package content_gc

import "gorm.io/gorm"

var (
	ShuttleCheckEndpoint = "/content/read/" // returns 404 if not available, 200 if content is available.
)

type BaseGC struct {
	DB *gorm.DB
}
