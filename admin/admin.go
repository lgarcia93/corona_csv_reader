package admin

import (
	"corona_csv_reader/models"
	"os"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"
)

var app *firebase.App

func init() {
	credentialsFile := os.Getenv("firebase_admin_credentials_file")

	opt := option.WithCredentialsFile(credentialsFile)
	app, _ = firebase.NewApp(context.Background(), nil, opt)
}

// InsertDailyStats ...
func InsertDailyStats(regions []models.Region) {
	ctx := context.Background()

	collectionKey := "daily_stats"

	client, _ := app.Firestore(ctx)

	iterator := client.Collection(collectionKey).Documents(ctx)

	allDocuments, _ := iterator.GetAll()

	for _, document := range allDocuments {
		document.Ref.Delete(ctx)
	}

	_, _ = client.Collection(collectionKey).NewDoc().Set(ctx, map[string][]models.Region{"data": regions})
}

// InsertTimeSeriesByRegion ...
func InsertTimeSeriesByRegion(dictionary map[string][]models.TimeSeriesItem) {

	ctx := context.Background()

	collectionKey := "time_series_by_region"

	client, _ := app.Firestore(ctx)

	iterator := client.Collection(collectionKey).Documents(ctx)

	allDocuments, _ := iterator.GetAll()

	for _, document := range allDocuments {
		document.Ref.Delete(ctx)
	}

	for key, value := range dictionary {
		_, _ = client.Collection(collectionKey).Doc(key).Set(ctx, map[string][]models.TimeSeriesItem{"data": value})
	}
}

// InsertTimeSeries ...
func InsertTimeSeries(timeSeries []models.TimeSeriesItem) {
	ctx := context.Background()

	client, _ := app.Firestore(ctx)

	collectionKey := "time_series"

	iterator := client.Collection(collectionKey).Documents(ctx)

	allDocuments, _ := iterator.GetAll()

	for _, document := range allDocuments {
		document.Ref.Delete(ctx)
	}

	_, _ = client.Collection(collectionKey).NewDoc().Set(ctx, map[string][]models.TimeSeriesItem{"data": timeSeries})
}
