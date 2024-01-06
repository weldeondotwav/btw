package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"github.com/weldeondotwav/btw/config"
)

var (
	ErrNoReminders = errors.New("no data in file to read")

	Config *config.AppConfig

	StartupDelay = time.Minute

	// https://github.com/microsoft/Windows-classic-samples/blob/44d192fd7ec6f2422b7d023891c5f805ada2c811/Samples/Win7Samples/begin/sdkdiff/sdkdiff.ico
	//go:embed assets/icon.ico
	iconData []byte

	DefaultRemindersFileContent = "# Lines starting with # or empty lines are ignored\n\nclean living room\ncheck mail"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// build our systray app
	systray.SetIcon(iconData)
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
				openFileWithDefault(Config.RemindersFilePath)
			case <-mEditConfig.ClickedCh:
				openFileWithDefault(config.Path())
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
	fmt.Println("btw closing")
}

// openFileWithDefault opens the file at path with the default application for that file type
func openFileWithDefault(path string) {
	openCmd := exec.Command("cmd", "/c", filepath.Clean(path))
	openCmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
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

	fileDataSplit := strings.Split(string(fileData), "\n")

	// filter out comments
	filteredLines := make([]string, 0)

	for _, v := range fileDataSplit {
		if len(v) < 1 {
			continue // ignore empty lines
		}

		if v[0] == '#' ||
			strings.TrimSpace(v) == "" {
			continue // ignore comments
		} else {
			filteredLines = append(filteredLines, v)
		}
	}
	return filteredLines, nil
}

// pick returns a random item from the input list choices
func pick(choices []string) string {
	i := rand.Intn(len(choices))
	return choices[i]
}

// Loads the user config, creating it if it doesn't exist
func loadConfig() {
	conf, err := config.Read()
	if err != nil {
		fmt.Println("ERROR: Failed to read config:", err)

		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("Creating default config...")

			defaultConfig := config.NewDefaultConfig()

			err = defaultConfig.Save()
			if err != nil {
				log.Fatal("Failed to save config! ", err)
			}

			conf = &defaultConfig

		} else {
			log.Fatal("Unhandled error while loading user config: ", err)
		}
	}

	fmt.Println("Loaded config")
	Config = conf // save the config object to our global var

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

		_, err = f.WriteString(DefaultRemindersFileContent)
		if err != nil {
			fmt.Println("ERROR: Failed to write default template to new reminders file: ", err)
		}
	}
}
