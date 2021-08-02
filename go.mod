module gitlab.com/gae4/trade-engine

go 1.16

require (
	github.com/emirpasic/gods v1.12.0
	github.com/gin-gonic/gin v1.7.2
	github.com/go-mysql-org/go-mysql v1.3.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/google/uuid v1.3.0
	github.com/jinzhu/gorm v1.9.16
	github.com/prometheus/common v0.2.0
	github.com/segmentio/kafka-go v0.4.17
	github.com/shopspring/decimal v1.2.0
	github.com/siddontang/go-log v0.0.0-20190221022429-1e957dd83bed
)

replace github.com/jinzhu/gorm v1.9.16 => github.com/jinzhu/gorm v1.9.10
