package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Category struct {
	Category string   `json:"category"`
	Answers  []string `json:"answers"`
}

func initiateCategories(categories *[]Category) {
	// Read the JSON file
	jsonFile, err := os.Open("answers.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			return
		}
	}(jsonFile)

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(byteValue, &categories)
	if err != nil {
		fmt.Println(err)
		return
	}
}
