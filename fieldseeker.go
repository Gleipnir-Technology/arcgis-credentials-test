package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type ArcGISItem struct {
	ID           string   `json:"id"`
	Owner        string   `json:"owner"`
	Created      int      `json:"created"`
	Modified     int      `json:"modified"`
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	URL          string   `json:"url"`
	Type         string   `json:"type"`
	TypeKeywords string   `json:"typeKeywords"`
	Description  string   `json:"description"`
	Tags         []string `json:"tags"`
	Snippet      string   `json:"snippet"`
}

type ArcGISSearchAggregation struct {
}
type ArcGISSearchResponse struct {
	Total             int                       `json:"total"`
	Start             int                       `json:"start"`
	Num               int                       `json:"num"`
	NextStart         int                       `json:"nextStart"`
	Results           []ArcGISItem              `json:"results"`
	Aggregations      []ArcGISSearchAggregation `json:"aggregations"`
	ServiceProperties []interface{}             `json:"servicePropertios"`
}

func findFieldseeker(access string) (*ArcGISSearchResponse, error) {
	baseURL := "https://www.arcgis.com/sharing/rest/search?q=FieldseekerGIS&f=pjson"
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		log.Printf("Failed to make request: %v", err)
		return nil, err
	}
	req.Header.Add("X-ESRI-Authorization", "Bearer "+access)
	client := http.Client{}
	log.Printf("GET %s", baseURL)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to do request: %v", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	log.Printf("Response %d", resp.StatusCode)
	if resp.StatusCode >= http.StatusBadRequest {
		if err != nil {
			return nil, fmt.Errorf("Got status code %d and failed to read response body: %v", resp.StatusCode, err)
		}
		bodyString := string(bodyBytes)
		var errorResp map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil {
			return nil, fmt.Errorf("API response JSON error: %d: %v", resp.StatusCode, errorResp)
		}
		return nil, fmt.Errorf("API returned error status %d: %s", resp.StatusCode, bodyString)
	}
	/*var content ArcGISSearchResponse
	err = json.Unmarshal(bodyBytes, &content)
	if err != nil {
		return nil, fmt.Errorf("Faied to unmarshal JSON: %v", err)
	}
	return &content, nil*/
	dest, err := os.Create("search.json")
	if err != nil {
		log.Printf("Faled to create output file: %v", err)
		return nil, err
	}
	_, err = io.Copy(dest, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("Faled to write output file: %v", err)
		return nil, err
	}
	log.Println("Wrote content to search.json")
	return nil, errors.New("not implemented")
}

func tryPortal(access string) {
	baseURL := "https://www.arcgis.com/sharing/rest/portals/self?f=pjson"
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		log.Printf("Failed to make request: %v", err)
		return
	}
	req.Header.Add("X-ESRI-Authorization", "Bearer "+access)
	client := http.Client{}
	log.Printf("GET %s", baseURL)
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to do request: %v", err)
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	log.Printf("Response %d", resp.StatusCode)
	dest, err := os.Create("portal.json")
	if err != nil {
		log.Printf("Faled to create output file: %v", err)
		return
	}
	_, err = io.Copy(dest, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("Faled to write output file: %v", err)
		return
	}
	log.Println("Wrote content to portal.json")
}
