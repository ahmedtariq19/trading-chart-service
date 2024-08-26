package db

import (
	"log"
	"time"
	"trading-chart-service/ohlc"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Candlestick struct {
	ID        uint      `gorm:"primaryKey"`
	Symbol    string    `gorm:"index"` // Indexed for faster querying
	Timestamp time.Time `gorm:"index"` // Indexed for faster querying
	Open      float64   `gorm:"type:decimal"`
	High      float64   `gorm:"type:decimal"`
	Low       float64   `gorm:"type:decimal"`
	Close     float64   `gorm:"type:decimal"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
type DB struct {
	conn *gorm.DB
}

func InitDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the Candlestick schema
	err = db.AutoMigrate(&Candlestick{})
	if err != nil {
		return nil, err
	}

	log.Println("Database connected and migrated")
	return db, nil
}

type Database struct {
	conn *gorm.DB
}

func NewDatabase(db *gorm.DB) *Database {
	return &Database{conn: db}
}

func (db *Database) StoreCandlestick(symbol string, cs ohlc.Candlestick) error {
	candle := Candlestick{
		Symbol:    symbol,
		Timestamp: cs.Timestamp,
		Open:      cs.Open,
		High:      cs.High,
		Low:       cs.Low,
		Close:     cs.Close,
	}

	result := db.conn.Create(&candle)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
