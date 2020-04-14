package database

import (
	"encoding/json"
	"github.com/google/logger"
	"os"
)

type Database struct {
	FileName string `long:"database.file" description:"File in which information about opened channels should be stored"`

	// Map between XUD identity public keys and an array of strings that contains the currencies on which a channel was opened
	channelsOpened map[string][]string

	file *os.File
}

func (database *Database) Init() {
	logger.Info("Starting database with file location: " + database.FileName)

	file, _ := os.OpenFile(database.FileName, os.O_RDWR, 0644)
	database.file = file

	logger.Info("Opening database file")

	if err := json.NewDecoder(database.file).Decode(&database.channelsOpened); err != nil {
		logger.Info("Could not open database file. Starting from scratch")
		database.channelsOpened = map[string][]string{}

		createdFile, err := os.Create(database.FileName)

		if err != nil {
			logger.Fatal("Could not create database file: " + err.Error())
		}

		database.file = createdFile

		database.write()
	}
}

func (database *Database) AddChannelsOpened(nodePubKey string, currency string) {
	database.channelsOpened[nodePubKey] = append(database.channelsOpened[nodePubKey], currency)
	database.write()
}

func (database *Database) GetChannelsOpened(nodePubKey string) []string {
	return database.channelsOpened[nodePubKey]
}

func (database *Database) write() {
	jsonMap, _ := json.MarshalIndent(database.channelsOpened, "", "  ")
	_, err := database.file.WriteAt(jsonMap, 0)

	if err != nil {
		logger.Error("Could not write database file: " + err.Error())
	}
}
