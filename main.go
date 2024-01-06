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
	"time"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/weldeondotwav/btw/config"
)

var (
	RemindersFilePath = "./reminders.txt"

	ErrNoReminders = errors.New("no data in file to read")

	Config *config.AppConfig

	StartupDelay = time.Minute

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
	mRemindNow := systray.AddMenuItem("Remind me now", "Sends an on-demand random reminder")
	mEditFile := systray.AddMenuItem("Open reminders file", "Opens the reminders file in the default text editor")
	mEditConfig := systray.AddMenuItem("Open app config", "Opens the application config in the default text editor")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Close the program")

	loadConfig()

	// event handlers (the way this works is so nice)
	go func() {
		for {
			select {
			case <-mRemindNow.ClickedCh:
				fmt.Println("Reminder requested!")
				sendRandomReminder()
			case <-mEditFile.ClickedCh:
				openRemindersFile()
			case <-mEditConfig.ClickedCh:
				openConfigFile()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	fmt.Println("Startup Delay: ", StartupDelay)
	time.Sleep(StartupDelay)

	go func() {
		for {
			sendRandomReminder()
			time.Sleep(Config.ReminderPeriod)
		}
	}()

}

func onExit() {
	fmt.Println("onExit")
}

// openRemindersFile opens the reminders file in the system default editor for .txt files
func openRemindersFile() {
	filePathAbs := filepath.Clean(Config.RemindersFilePath)
	openCmd := exec.Command("cmd", "/c", filePathAbs)
	openCmd.Start()
}

// openConfigFile opens the application config file in the system default editor for .json files
func openConfigFile() {
	openCmd := exec.Command("cmd", "/c", config.Path())
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
	fileData, err := os.ReadFile(Config.RemindersFilePath)

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
		if len(v) < 1 {
			continue
		}

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

// Loads the user config, or creates one if it doesn't exist
func loadConfig() {
	conf, err := config.Read()
	if err != nil {
		fmt.Println("ERROR: Failed to read config:", err)

		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Creating default config...")

			exConfig := config.NewDefaultConfig()

			err = exConfig.Save()
			if err != nil {
				log.Fatal("Failed to save config! ", err)
			}

			conf = &exConfig

		} else {
			log.Fatal("Unhandled error while loading user config: ", err)
		}
	}

	fmt.Println("Loaded config")
	Config = conf

	// Also create the reminders file if it doesn't exist
	_, err = os.Stat(conf.RemindersFilePath)
	if err != nil {
		fmt.Println("Error when checking for reminders file", err)
		// need to check the err first frfr
		fmt.Println("Reminders file not found, creating it at ", conf.RemindersFilePath)

		f, err := os.Create(conf.RemindersFilePath)
		if err != nil {
			log.Fatal("Failed to create reminders file: ", err)
		}
		defer f.Close()

		_, err = f.WriteString("# Lines starting with # or empty lines are ignored\n\nclean living room\ncheck mail")
		if err != nil {
			fmt.Println("ERROR: Failed to write default template to new reminders file: ", err)
		}
	}
}
