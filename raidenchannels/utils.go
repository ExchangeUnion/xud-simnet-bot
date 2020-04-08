package raidenchannels

import (
	"encoding/gob"
	"os"
)

func saveInactiveTimes(dataPath string) {
	file, err := os.OpenFile(dataPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer file.Close()

	if err != nil {
		log.Warning("Could not write channel data to disk: %v", err)
		return
	}

	encoder := gob.NewEncoder(file)
	encoder.Encode(inactiveTimes)
}

func readInactiveTimes(dataPath string) {
	if _, err := os.Stat(dataPath); err != nil {
		// File does not exist
		return
	}

	file, err := os.Open(dataPath)
	defer file.Close()

	if err != nil {
		log.Warning("Could not read channel data from disk: %v", err)
		return
	}

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&inactiveTimes)

	if err != nil {
		log.Warning("Could not parse channel data from disk: %v", err)
	}
}
