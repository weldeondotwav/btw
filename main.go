package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

var (
	RemindersFilePath = "./reminders.txt"

	ErrNoReminders = errors.New("no data in file to read")

	// notificationFrequency = time.Second * 10
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {

	// https://github.com/microsoft/Windows-classic-samples/blob/44d192fd7ec6f2422b7d023891c5f805ada2c811/Samples/Win7Samples/begin/sdkdiff/sdkdiff.ico
	ico, err := os.ReadFile("assets/icon.ico")
	if err != nil {
		log.Fatal("Failed to read icon: ", err)
	}

	// build our systray app

	systray.SetIcon(ico)
	systray.SetTitle("btw")
	systray.SetTooltip("btw: reminders")
	mQuit := systray.AddMenuItem("Quit", "Close the program")
	mRemindNow := systray.AddMenuItem("Remind me now", "Sends an on-demand random reminder")
	mEditFile := systray.AddMenuItem("Open reminders file", "Opens the reminders file in the default text editor")

	// event handlers (the way this works is so nice)
	go func() {
		for {
			select {
			case <-mRemindNow.ClickedCh:
				fmt.Println("Reminder requested!")
				sendRandomReminder()
			case <-mEditFile.ClickedCh:
				openRemindersFile()
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				fmt.Println("Finished quitting")
				return
			}
		}

	}()
}

func onExit() {
	fmt.Println("onExit")
}

// openRemindersFile opens the reminders file in the system default editor for .txt files
func openRemindersFile() {
	filePathAbs, err := filepath.Abs(RemindersFilePath)
	if err != nil {
		log.Fatal("Failed to find absolute path of reminders file")
	}

	openCmd := exec.Command("cmd", "/c", filePathAbs)

	openCmd.Start()

}

// sendRandomReminder selects a random item from the users reminders list and sends it as a tray notification
func sendRandomReminder() {
	reminders, err := getReminders()
	if err != nil {
		log.Fatal("Failed to read reminders: ", err)
	}

	reminderToShow := pick(reminders)

	fmt.Println("showing reminder: ", reminderToShow)

	err = beeep.Notify("btw", reminderToShow, "")
	if err != nil {
		fmt.Println("ERROR: Failed to send reminder: ", err)
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
	// filter out comments
	filteredLines := make([]string, 0)

	for _, v := range fileDataSplit {
		if v[0] == '#' ||
			strings.TrimSpace(v) == "" {
			continue
		} else {
			filteredLines = append(filteredLines, v)
		}
	}

	return filteredLines, nil
}

// pick picks a random item from the input list choices
func pick(choices []string) string {
	i := rand.Intn(len(choices))

	return choices[i]
}
