package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sod-auctions/auctions-db"
	"log"
	"net/http"
	"os"
	"strconv"
)

type ErrorMessage struct {
	Error string `json:"error"`
}

var database *auctions_db.Database

func init() {
	log.SetFlags(0)
	var err error
	database, err = auctions_db.NewDatabase(os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
}

type PriceDistribution struct {
	BuyoutEach int32 `json:"buyoutEach"`
	Quantity   int32 `json:"quantity"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	realmId, _ := strconv.Atoi(event.QueryStringParameters["realmId"])
	auctionHouseId, _ := strconv.Atoi(event.QueryStringParameters["auctionHouseId"])
	itemId, _ := strconv.Atoi(event.QueryStringParameters["itemId"])

	priceDistributions, err := database.GetPriceDistributions(int16(realmId), int16(auctionHouseId), int32(itemId))
	if err != nil {
		log.Printf("An error occurred: %v\n", err)

		errorMessage := ErrorMessage{Error: "An internal error occurred"}
		body, _ := json.Marshal(errorMessage)

		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type":                 "application/json",
				"Access-Control-Allow-Origin":  "http://localhost:3000",
				"Access-Control-Allow-Methods": "GET, OPTIONS",
				"Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept, Authorization",
			},
			Body: string(body),
		}, nil
	}

	var vPriceDistributions []*PriceDistribution
	for _, item := range priceDistributions {
		vPriceDistributions = append(vPriceDistributions, &PriceDistribution{
			BuyoutEach: item.BuyoutEach,
			Quantity:   item.Quantity,
		})
	}

	body, _ := json.Marshal(vPriceDistributions)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "http://localhost:3000",
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Origin, X-Requested-With, Content-Type, Accept, Authorization",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
