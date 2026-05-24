package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectTestDatabase(t *testing.T) *mongo.Database {
	t.Helper()

	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/auctions?authSource=admin"
	}

	mongoClient, connectionError := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURL))
	if connectionError != nil {
		t.Skipf("skipping integration test: cannot connect to MongoDB: %v", connectionError)
	}

	pingError := mongoClient.Ping(context.Background(), nil)
	if pingError != nil {
		t.Skipf("skipping integration test: cannot ping MongoDB: %v", pingError)
	}

	return mongoClient.Database("auctions_test")
}

func TestAuctionAutoClose(t *testing.T) {
	testDatabase := connectTestDatabase(t)
	defer testDatabase.Collection("auctions").Drop(context.Background())

	os.Setenv("AUCTION_DURATION", "1s")
	defer os.Unsetenv("AUCTION_DURATION")

	auctionRepository := NewAuctionRepository(testDatabase)

	auctionEntity, createError := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"A test product for auction auto-close verification",
		auction_entity.New,
	)
	if createError != nil {
		t.Fatalf("failed to create auction entity: %v", createError)
	}

	insertError := auctionRepository.CreateAuction(context.Background(), auctionEntity)
	if insertError != nil {
		t.Fatalf("failed to insert auction: %v", insertError)
	}

	if auctionEntity.Status != auction_entity.Active {
		t.Fatalf("expected auction status Active after creation, got %v", auctionEntity.Status)
	}

	time.Sleep(2 * time.Second)

	var auctionDocument bson.M
	findError := auctionRepository.Collection.FindOne(
		context.Background(),
		bson.M{"_id": auctionEntity.Id},
	).Decode(&auctionDocument)
	if findError != nil {
		t.Fatalf("failed to find auction after expiry: %v", findError)
	}

	actualStatus := int(auctionDocument["status"].(int32))
	expectedStatus := int(auction_entity.Completed)

	if actualStatus != expectedStatus {
		t.Errorf("expected auction status %d (Completed), got %d", expectedStatus, actualStatus)
	}
}

func TestAuctionDoesNotCloseBeforeExpiry(t *testing.T) {
	testDatabase := connectTestDatabase(t)
	defer testDatabase.Collection("auctions").Drop(context.Background())

	os.Setenv("AUCTION_DURATION", "60s")
	defer os.Unsetenv("AUCTION_DURATION")

	auctionRepository := NewAuctionRepository(testDatabase)

	auctionEntity, createError := auction_entity.CreateAuction(
		"Long Running Product",
		"Electronics",
		"A test product that should remain active during the test",
		auction_entity.New,
	)
	if createError != nil {
		t.Fatalf("failed to create auction entity: %v", createError)
	}

	insertError := auctionRepository.CreateAuction(context.Background(), auctionEntity)
	if insertError != nil {
		t.Fatalf("failed to insert auction: %v", insertError)
	}

	var auctionDocument bson.M
	findError := auctionRepository.Collection.FindOne(
		context.Background(),
		bson.M{"_id": auctionEntity.Id},
	).Decode(&auctionDocument)
	if findError != nil {
		t.Fatalf("failed to find auction: %v", findError)
	}

	actualStatus := int(auctionDocument["status"].(int32))
	expectedStatus := int(auction_entity.Active)

	if actualStatus != expectedStatus {
		t.Errorf("expected auction status %d (Active), got %d", expectedStatus, actualStatus)
	}
}

func TestGetAuctionDuration(t *testing.T) {
	os.Setenv("AUCTION_DURATION", "30s")
	defer os.Unsetenv("AUCTION_DURATION")

	auctionDuration := getAuctionDuration()
	if auctionDuration != 30*time.Second {
		t.Errorf("expected 30s, got %v", auctionDuration)
	}
}

func TestGetAuctionDurationDefaultsToFiveMinutes(t *testing.T) {
	os.Unsetenv("AUCTION_DURATION")

	auctionDuration := getAuctionDuration()
	if auctionDuration != 5*time.Minute {
		t.Errorf("expected default 5m, got %v", auctionDuration)
	}
}
