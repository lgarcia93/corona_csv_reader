package core

import (
	"bufio"
	"corona_csv_reader/admin"
	"corona_csv_reader/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var regions = []models.Region{}

var openCageAPIKey = os.Getenv("opencageapikey")

var url string = "https://api.opencagedata.com/geocode/v1/json?key=%s&q=%s"

func getCoords(placename string) models.LatLng {

	formattedPlace := strings.Replace(placename, " ", "+", 30)

	resp, err := http.Get(fmt.Sprintf(url, openCageAPIKey, formattedPlace))

	if err != nil {
		return models.LatLng{Latitude: 0.0, Longitude: 0.0}
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	var geocodingData models.GeocodingResponse

	json.Unmarshal(bodyBytes, &geocodingData)

	if len(geocodingData.Results) > 0 {
		return geocodingData.Results[0].Coords
	}

	return models.LatLng{Latitude: 0.0, Longitude: 0.0}
}

func getCoordsForRegions(regions []models.Region) []models.Region {
	outputRegion := []models.Region{}

	for _, region := range regions {
		fmt.Printf("Getting coordinates for %s %s \n", region.Province, region.Country)
		coords := getCoords(fmt.Sprintf("%s, %s", region.Province, region.Country))

		region.Latitude = coords.Latitude
		region.Longitude = coords.Longitude

		outputRegion = append(outputRegion, region)
	}

	return outputRegion
}

func getCSVFileNamesList() []string {
	var files []string

	root := os.Getenv("csv_dir")

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".csv" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return files
}

func getDateFromFileName(fileName string) (time.Time, error) {
	extension := filepath.Ext(fileName)

	filePathWithoutExtension := fileName[0 : len(fileName)-len(extension)]

	fileNameWithoutPathParts := strings.Split(filePathWithoutExtension, "/")

	fileNameWithoutPath := fileNameWithoutPathParts[len(fileNameWithoutPathParts)-1]

	layout := "01-02-2006"

	return time.Parse(layout, fileNameWithoutPath)
}

func processCSVLine(csvLine string) string {
	outputLine := strings.Replace(csvLine, "US", "United States", -1)
	outputLine = strings.Replace(outputLine, "Mainland China", "China", -1)
	outputLine = strings.Replace(outputLine, "\"", "", -1)
	outputLine = strings.Replace(outputLine, ", ", " ", -1)

	return outputLine
}

func getRegionsFromCSV(fileName string) []models.Region {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_regions := []models.Region{}

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		if count != 0 {

			processedLine := processCSVLine(scanner.Text())

			splitted := strings.Split(processedLine, ",")

			confirmedCases, _ := strconv.Atoi(splitted[3])
			confirmedDeaths, _ := strconv.Atoi(splitted[4])

			recovered, _ := strconv.Atoi(splitted[5])

			region := models.Region{
				Province:        strings.Replace(splitted[0], "\"", "", 15),
				Country:         splitted[1],
				LastUpdated:     splitted[2],
				ConfirmedCases:  confirmedCases,
				ConfirmedDeaths: confirmedDeaths,
				Recovered:       recovered,
			}

			if region.Country == "Macau" {
				region.Country = "China"
			}

			region.Key = generateRegionKey(region)

			_regions = append(_regions, region)
		}
		count++
	}

	return _regions
}

func generateRegionKey(region models.Region) string {
	return strings.Replace(fmt.Sprintf("%s%s", region.Province, region.Country), " ", "", 10)
}

func convertAndSaveToJSON(outputName string, object interface{}) {
	jsonBytes, _ := json.Marshal(object)

	ioutil.WriteFile(outputName, jsonBytes, 0644)
}

func GenerateTimeSeries() {
	fileNames := getCSVFileNamesList()

	dataDict := make(map[string][]models.TimeSeriesItem)

	timeSeriesList := []models.TimeSeriesItem{}

	currentDayRegions := []models.Region{}

	initialDate := time.Date(1970, 01, 01, 00, 00, 00, 00, time.UTC)

	for _, fileName := range fileNames {

		fmt.Printf("Processing %s \n", fileName)

		fileDate, _ := getDateFromFileName(fileName)

		regionList := getRegionsFromCSV(fileName)

		if fileDate.Unix() > initialDate.Unix() {
			currentDayRegions = regionList
		}

		dayTotalCases := 0
		dayTotalDeaths := 0
		dayTotalRecoveries := 0

		for _, region := range regionList {

			regionKey := generateRegionKey(region)

			if dataDict[regionKey] == nil {
				dataDict[regionKey] = []models.TimeSeriesItem{}
			}

			regionTimeSeriesItem := models.TimeSeriesItem{
				Date:            fileDate,
				ConfirmedCases:  region.ConfirmedCases,
				ConfirmedDeaths: region.ConfirmedDeaths,
				Recoveries:      region.Recovered,
			}

			dataDict[regionKey] = append(dataDict[regionKey], regionTimeSeriesItem)

			dayTotalCases += region.ConfirmedCases
			dayTotalDeaths += region.ConfirmedDeaths
			dayTotalRecoveries += region.Recovered
		}

		timeSeriesItem := models.TimeSeriesItem{
			Date:            fileDate,
			ConfirmedCases:  dayTotalCases,
			ConfirmedDeaths: dayTotalDeaths,
			Recoveries:      dayTotalRecoveries,
		}

		timeSeriesList = append(timeSeriesList, timeSeriesItem)
	}

	fmt.Print("Fetching coordinates... \n")

	currentDayRegions = getCoordsForRegions(currentDayRegions)

	convertAndSaveToJSON("corona_time_series.json", timeSeriesList)
	convertAndSaveToJSON("corona_time_series_by_region.json", dataDict)
	convertAndSaveToJSON("current_day_stats.json", currentDayRegions)

	admin.InsertDailyStats(currentDayRegions)
	admin.InsertTimeSeriesByRegion(dataDict)
	admin.InsertTimeSeries(timeSeriesList)
}
