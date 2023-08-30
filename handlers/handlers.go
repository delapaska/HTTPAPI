package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/delapaska/AvitoTest/db/segments"
)

type Segm struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
	TTL  string `json:"ttl"`
}

type Percent struct {
	Percent int `json:"percent"`
}

type History struct {
	ID        int    `json:"id"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type HistoryEntry struct {
	ID      int    `json:"id"`
	Segment string `json:"segment"`
	Action  string `json:"action"`
	Date    string `json:"date"`
}

func CreateHandler(rep segments.SegmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user Segm
		json.NewDecoder(r.Body).Decode(&user)
		rep.CreateSegment(user.Name)

	}

}

func DeleteHandler(rep segments.SegmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user Segm
		json.NewDecoder(r.Body).Decode(&user)
		rep.DeleteSegment(user.Name)

	}

}

func AddUserHandler(rep segments.SegmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user Segm
		json.NewDecoder(r.Body).Decode(&user)
		rep.AddUserSegment(user.ID, user.Name, user.TTL)

	}

}

func DeleteUserHandler(rep segments.SegmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user Segm
		json.NewDecoder(r.Body).Decode(&user)
		rep.DeleteUserSegment(user.ID, user.Name)

	}

}

func ReturnSegmentHandler(rep segments.SegmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user Segm
		json.NewDecoder(r.Body).Decode(&user)

		var segments []string
		rows := rep.ReturnSegment(user.ID)
		for rows.Next() {
			var segment string
			err := rows.Scan(&segment)
			if err != nil {
				http.Error(w, "Result scanning error", http.StatusInternalServerError)
				return
			}
			segments = append(segments, segment)
		}

		jsonData, err := json.Marshal(segments)
		if err != nil {
			http.Error(w, "JSON marshaling error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)

	}
}

var URL string

func GetUserHistoryHandler(rep segments.SegmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var hist History
		json.NewDecoder(r.Body).Decode(&hist)

		fmt.Println(hist.ID, hist.StartDate, hist.EndDate)
		var history []HistoryEntry

		rows := rep.GetUserHistory(hist.ID, hist.StartDate, hist.EndDate)
		for rows.Next() {
			var entry HistoryEntry
			err := rows.Scan(&entry.ID, &entry.Segment, &entry.Action, &entry.Date)
			if err != nil {
				http.Error(w, "Result scanning error", http.StatusInternalServerError)
				return
			}
			history = append(history, entry)
		}


		filename := fmt.Sprintf("user_history_%d.csv", hist.ID)
		log.Printf("Creating CSV file: %s", filename)

		file, err := os.Create(filename)
		if err != nil {
			log.Printf("Error creating file: %v", err)
			http.Error(w, "File creation error", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		csvWriter := csv.NewWriter(file)
		defer csvWriter.Flush()


		headers := []string{"ID", "Segment", "Action", "Date"}
		csvWriter.Write(headers)

		for _, entry := range history {

			csvWriter.Write([]string{
				fmt.Sprintf("%d", entry.ID),
				entry.Segment,
				entry.Action,
				entry.Date,
			})
		}

		downloadURL := fmt.Sprintf("http://localhost:8080/download/%s", filename)
		log.Printf("CSV file created. Download URL: %s", downloadURL)
		fmt.Fprintf(w, "CSV file created. You can download it from: %s", downloadURL)
		URL = downloadURL

	}
}

func DownloadCSVHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(URL, "/")
	file_N := parts[len(parts)-1]
	
	fileName := file_N
	filePath := filepath.Join(".", fileName)

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "text/csv")

	_, err = io.Copy(w, file)

	if err != nil {
		http.Error(w, "Error while copying file content", http.StatusInternalServerError)
		return
	}

}
func DistributeUsersHandler(rep segments.SegmentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var per Percent
		json.NewDecoder(r.Body).Decode(&per)
		var users []int
		rows := rep.DistributeUsers()

		for rows.Next() {
			var user int
			err := rows.Scan(&user)

			if err != nil {
				http.Error(w, "Result scanning error", http.StatusInternalServerError)
				return
			}

			users = append(users, user)
		}

		rows = rep.GetSegments()
		var segments []string
		for rows.Next() {
			var segment string
			err := rows.Scan(&segment)

			if err != nil {
				http.Error(w, "Result scanning error", http.StatusInternalServerError)
				return
			}

			segments = append(segments, segment)
		}
		percentage := len(users) * per.Percent / 100
		rand.Seed(time.Now().UnixNano())
		for _, segment := range segments {
			shuffledSlice := make([]int, len(users))
			copy(shuffledSlice, users)

			shuffleSlice(shuffledSlice)
			for i := 0; i < percentage; i++ {
				randomValue := shuffledSlice[0]
				rep.AddUserSegment(randomValue, segment, "")
				shuffledSlice = shuffledSlice[1:]

				fmt.Println(i)
			}
		}

		fmt.Println(users)
		fmt.Println(segments)
	}
}
func shuffleSlice(slice []int) {
	rand.Seed(time.Now().UnixNano()) 

	for i := len(slice) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
