package main

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

var (
	RemindersFilePath = "./reminders.txt"

	ErrNoReminders = errors.New("no data in file to read")

	notificationFrequency = time.Second * 10
)

func main() {

	for {

		reminders, err := getReminders()
		if err != nil {
			log.Fatal("Failed to read reminders: ", err)
		}

		reminderToShow := pickReminder(reminders)

		err = beeep.Notify("btw", reminderToShow, "")
		if err != nil {
			panic(err)
		}


		time.Sleep(notificationFrequency)
	}
}

// getReminders returns all the lines in the reminders file as a string array
func getReminders() ([]string, error) {
	fileData, err := os.ReadFile(RemindersFilePath)

	if err != nil {
		return nil, err
	}

	if len(fileData) == 0 {
		return nil, ErrNoReminders
	}

	fileDataString := string(fileData)

	fileDataSplit := strings.Split(fileDataString, "\n")

	return fileDataSplit, nil
}

func pickReminder(choices []string) string {
	i := rand.Intn(len(choices))

	return choices[i]
}
