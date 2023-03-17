package twitch

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/lleadbet/chat-controller/config"
	"github.com/micmonay/keybd_event"
	"go.uber.org/zap"
)

// a message key maps to one or more key presses
var messageMap = make(map[string]*config.ChatMessageConfig)

func ChatReader(logger *zap.Logger) {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}

	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer watcher.Close()

	go watcherFunc(logger, watcher)
	go watcher.Add("config.yaml")
	config, err := config.NewConfig(logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	err = configUpdate(logger, config)
	if err != nil {
		logger.Fatal(err.Error())
	}
	client := twitch.NewAnonymousClient()

	logger.Info("Connecting to Twitch chat...")
	client.Join(config.Username)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		m := strings.TrimSpace(strings.ToLower(message.Message))
		if messageMap[m] != nil {
			logger.Info(fmt.Sprintf("Chat command found for %v", m))
			config := messageMap[m]
			for _, key := range config.Key {
				key = strings.ToUpper(key)

				if key == "SHIFT" {
					kb.HasSHIFT(true)
				} else if key == "CONTROL" {
					kb.HasCTRL(true)
				} else if key == "ALT" {
					kb.HasALT(true)
				} else {
					kb.AddKey(KeyboardEvents[key])
				}
			}
			kb.Press()
			logger.Debug("Pressing keys...")
			logger.Debug(fmt.Sprint(time.Duration(config.Duration * float64(time.Second))))
			time.Sleep(time.Duration(config.Duration) * time.Second)
			logger.Debug("Done!")
			kb.Release()
		}
	})

	client.OnConnect(func() {
		logger.Info(fmt.Sprintf("Successfully connected to #%v", config.Username))
	})

	go func() {
		err := client.Connect()
		println(err.Error())
		if err != nil {
			panic(err)
		}
	}()
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel
	//time for cleanup before exit
	fmt.Println("Thanks for using this tool!")
}

func configUpdate(logger *zap.Logger, c *config.Config) error {
	messageMap = map[string]*config.ChatMessageConfig{}
	for _, cm := range c.ChatMessage {
		for _, message := range cm.Message {
			if messageMap[message] != nil {
				logger.Fatal(fmt.Sprintf("Command already exists '%v'", message))
			}
			for i, key := range cm.Key {
				cm.Key[i] = strings.ToUpper(key)
				key = cm.Key[i]
				if key == "SHIFT" || key == "CONTROL" || key == "ALT" {
					continue
				} else if KeyboardEvents[key] == keybd_event.VK_RESERVED {
					return errors.New(fmt.Sprintf("Invalid key provided %v", key))
				}
			}
			messageMap[message] = &config.ChatMessageConfig{Key: cm.Key, Duration: cm.Duration}
		}
	}
	return nil
}

func watcherFunc(logger *zap.Logger, watcher *fsnotify.Watcher) {
	var (
		// Wait 100ms for new events; each new event resets the timer.
		waitFor = 100 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)

		// Callback we run.
		printEvent = func(e fsnotify.Event) {
			logger.Info("Configuration updated")
			config, err := config.NewConfig(logger)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			err = configUpdate(logger, config)
			if err != nil {
				logger.Error(err.Error())
			}
			// Don't need to remove the timer if you don't have a lot of files.
			mu.Lock()
			delete(timers, e.Name)
			mu.Unlock()
		}
	)
	for {
		select {
		case e, ok := <-watcher.Events:
			if !ok {
				return
			}
			if e.Has(fsnotify.Write) {
				// Get timer.
				mu.Lock()
				t, ok := timers[e.Name]
				mu.Unlock()

				// No timer yet, so create one.
				if !ok {
					t = time.AfterFunc(math.MaxInt64, func() { printEvent(e) })
					t.Stop()

					mu.Lock()
					timers[e.Name] = t
					mu.Unlock()
				}

				// Reset the timer for this path, so it will start from 100ms again.
				t.Reset(waitFor)
			}
		}
	}
}
