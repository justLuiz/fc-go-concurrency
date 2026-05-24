package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}

type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (auctionRepository *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := auctionRepository.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	go auctionRepository.closeAuctionAfterExpiry(auctionEntity.Id, getAuctionDuration())

	return nil
}

func getAuctionDuration() time.Duration {
	auctionDurationEnv := os.Getenv("AUCTION_DURATION")
	auctionDuration, parseError := time.ParseDuration(auctionDurationEnv)
	if parseError != nil {
		return 5 * time.Minute
	}
	return auctionDuration
}

func (auctionRepository *AuctionRepository) closeAuctionAfterExpiry(auctionID string, auctionDuration time.Duration) {
	time.Sleep(auctionDuration)

	updateFilter := bson.M{"_id": auctionID}
	updateDocument := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	updateContext, cancelUpdateContext := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelUpdateContext()

	_, updateError := auctionRepository.Collection.UpdateOne(updateContext, updateFilter, updateDocument)
	if updateError != nil {
		logger.Error("Error trying to close auction after expiry", updateError)
	}
}
