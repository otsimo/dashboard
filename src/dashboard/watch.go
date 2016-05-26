package dashboard

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsnotify/fsnotify"
)

func watchFile(configFilePath string, s *Server) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op == fsnotify.Write {
					log.Println("modified file:", event.Name)
					s.RereadConfig()
				} else if event.Op == fsnotify.Remove {
					watcher.Remove(configFilePath)
					watcher.Add(configFilePath)
					s.RereadConfig()
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(configFilePath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
