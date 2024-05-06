package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/joho/godotenv"
)

func getEnvVar(key string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return os.Getenv(key)
}
func getVersions() []*version.Version {

	fmt.Printf("getting versions from Content Hub One\n")

	jsonMapInstance := map[string]string{
		"query": `
			{
				allRsjssversion {
					results {
						id
						version
					}
				}
			}`,
	}

	jsonResult, err := json.Marshal(jsonMapInstance)
	if err != nil {
		fmt.Printf("There was an error marshaling the JSON instance %v\n", err)
	}

	newRequest, _ := http.NewRequest(
		"POST",
		"https://edge.sitecorecloud.io/api/graphql/v1",
		bytes.NewBuffer(jsonResult))

	apiKey := getEnvVar("CONTENT_API_KEY")
	fmt.Printf("apiKey: %s\n", apiKey)

	newRequest.Header.Set("Content-Type", "application/json")
	newRequest.Header.Set("X-GQL-Token", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(newRequest)
	if err != nil {
		fmt.Printf("There was an error executing the request%v\n", err)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Data Read Error%v\n", err)
	}

	fmt.Printf("responseData: %s\n", string(responseData))

	var responseMap map[string]interface{}
	err = json.Unmarshal(responseData, &responseMap)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	results := responseMap["data"].(map[string]interface{})["allRsjssversion"].(map[string]interface{})["results"].([]interface{})

	versions := []*version.Version{}

	for _, result := range results {
		version, _ := version.NewVersion(result.(map[string]interface{})["version"].(string))
		versions = append(versions, version)
	}

	sort.Sort(sort.Reverse(version.Collection(versions)))
	return versions
}

func getInstructions() []UpgradeStep {

	fmt.Printf("getting instructions from Content Hub One\n")

	jsonMapInstance := map[string]string{
		"query": `
		{
			allRsjssversion(
				where: {
					version_eq: "22.0.0"
				}
			) {
				results {
				id
				name
				version
				instructionsmd
				}
			}
		}`,
	}

	jsonResult, err := json.Marshal(jsonMapInstance)
	if err != nil {
		fmt.Printf("There was an error marshaling the JSON instance %v\n", err)
	}

	newRequest, _ := http.NewRequest(
		"POST",
		"https://edge.sitecorecloud.io/api/graphql/v1",
		bytes.NewBuffer(jsonResult))

	apiKey := getEnvVar("CONTENT_API_KEY")
	fmt.Printf("apiKey: %s\n", apiKey)

	newRequest.Header.Set("Content-Type", "application/json")
	newRequest.Header.Set("X-GQL-Token", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(newRequest)
	if err != nil {
		fmt.Printf("There was an error executing the request%v\n", err)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Data Read Error%v\n", err)
	}

	fmt.Printf("responseData: %s\n", string(responseData))

	var responseMap map[string]interface{}
	err = json.Unmarshal(responseData, &responseMap)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	results := responseMap["data"].(map[string]interface{})["allRsjssversion"].(map[string]interface{})["results"].([]interface{})

	upgradeSteps := []UpgradeStep{}

	for _, result := range results {
		upgradeStep := UpgradeStep{
			Id:      result.(map[string]interface{})["id"].(string),
			Version: result.(map[string]interface{})["version"].(string),
			Steps:   Markdown(result.(map[string]interface{})["instructionsmd"].(string)),
		}
		upgradeSteps = append(upgradeSteps, upgradeStep)
	}

	return upgradeSteps
}
