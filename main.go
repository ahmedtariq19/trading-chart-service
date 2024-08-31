package main

import (
	"log"
	"net"

	"trading-chart-service/binance"
	"trading-chart-service/db"
	"trading-chart-service/ohlc"
	pb "trading-chart-service/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedCandlestickServiceServer
	aggregator *ohlc.Aggregator
	db         *db.Database
}

func (s *server) StreamCandlesticks(req *pb.Empty, stream pb.CandlestickService_StreamCandlesticksServer) error {
	candlestickChan := s.aggregator.Subscribe()

	for candlestick := range candlestickChan {
		cs := &pb.Candlestick{
			Symbol:    candlestick.Symbol,
			Timestamp: candlestick.Timestamp.Unix(),
			Open:      candlestick.Open,
			High:      candlestick.High,
			Low:       candlestick.Low,
			Close:     candlestick.Close,
		}

		if err := stream.Send(cs); err != nil {
			return err
		}
	}
	return nil
}

func main() {

	dsn := "host=localhost user=user password=password dbname=database port=5432 sslmode=allow"
	databaseConn, err := db.InitDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	database := db.NewDatabase(databaseConn)

	// Initialize the OHLC aggregator
	aggregator := ohlc.NewAggregator()

	// Start fetching ticks from Binance
	tickChan := make(chan binance.Tick)
	symbols := []string{"btcusdt", "ethusdt", "pepeusdt"}
	go binance.Connect(symbols, tickChan)

	// Process ticks and handle OHLC aggregation and database storage
	go func() {
		for tick := range tickChan {
			candlestick, isComplete := aggregator.ProcessTick(tick)
			if isComplete {
				// Store the completed candlestick in the database
				if err := database.StoreCandlestick(tick.Symbol, candlestick); err != nil {
					log.Printf("Failed to store candlestick in database: %v", err)
				}
			}
		}
	}()

	// Start the gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterCandlestickServiceServer(s, &server{
		aggregator: aggregator,
		db:         database,
	})

	log.Printf("Server is running on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
