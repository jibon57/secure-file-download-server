package internal

import (
	"log"
	"os"
	"path"
	"time"
)

func StartScheduler(shutdown chan os.Signal) {
	hourlyChecker := time.NewTicker(1 * time.Hour)
	defer hourlyChecker.Stop()

	for {
		select {
		case <-shutdown:
			log.Println("Stopping scheduler")
			return
		case <-hourlyChecker.C:
			checkDelFileBackupPath()
		}
	}
}

func checkDelFileBackupPath() {
	if !AppCnf.EnableDelFileBackup {
		// nothing to do
		return
	}

	checkTime := time.Now().Add(-AppCnf.DelFileBackupDuration)
	entries, err := os.ReadDir(AppCnf.DelFileBackupPath)
	if err != nil {
		log.Println(err)
		return
	}
	for _, et := range entries {
		if et.IsDir() {
			continue
		}
		info, err := et.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(checkTime) {
			// we can remove this file
			fileToDelete := path.Join(AppCnf.DelFileBackupPath, et.Name())
			log.Println("deleting file:", fileToDelete, "because of created", checkTime, "which is older than", AppCnf.DelFileBackupDuration)

			err = os.Remove(fileToDelete)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
