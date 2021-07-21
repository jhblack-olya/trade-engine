module gitlab.com/gae4/trade-engine

go 1.16

require (
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/prometheus/common v0.2.0
	github.com/segmentio/kafka-go v0.4.17
	github.com/shopspring/decimal v1.2.0
	github.com/siddontang/go-log v0.0.0-20190221022429-1e957dd83bed
)

replace github.com/jinzhu/gorm v1.9.16 => github.com/jinzhu/gorm v1.9.10