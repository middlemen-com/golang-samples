// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Sample image-processing is a Cloud Run service which performs asynchronous processing on images.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/GoogleCloudPlatform/golang-samples/run/image-processing/imagemagick"
)

func main() {
	http.HandleFunc("/", HelloPubSub)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

// [START run_imageproc_controller]

// PubSubMessage is the payload of a Pub/Sub event.
type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// HelloPubSub consumes a Pub/Sub message.
func HelloPubSub(w http.ResponseWriter, r *http.Request) {
	// Parse the Pub/Sub message.
	var m PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Printf("json.NewDecoder: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var e imagemagick.GCSEvent
	dataReader := bytes.NewReader(m.Message.Data)
	if err := json.NewDecoder(dataReader).Decode(&e); err != nil {
		log.Printf("json.NewDecoder: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if e.Name == "" || e.Bucket == "" {
		log.Printf("invalid GCSEvent: expected name and bucket")
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	err := imagemagick.BlurOffensiveImages(r.Context(), e)
	if err != nil {
		log.Printf("imagemagick.BlurOffensiveImages: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// [END run_imageproc_controller]
