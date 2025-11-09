package internal

import (
	"os"
	"path"
	"time"
)

type AppConfig struct {
	Port                  int64         `yaml:"port"`
	ApiKey                string        `yaml:"api_key"`
	ApiSecret             string        `yaml:"api_secret"`
	Path                  string        `yaml:"path"`
	NginxFileServePath    string        `yaml:"nginx_file_serve_path"`
	Debug                 bool          `yaml:"debug"`
	Compress              bool          `yaml:"compress"`
	DeleteEmptyDir        bool          `yaml:"delete_empty_dir"`
	EnableDelFileBackup   bool          `yaml:"enable_del_file_backup"`
	DelFileBackupPath     string        `yaml:"del_file_backup_path"`
	DelFileBackupDuration time.Duration `yaml:"del_file_backup_duration"`
}

var AppCnf AppConfig

func CreateOrUpdateDirs() error {
	err := os.MkdirAll(AppCnf.Path, 0755)
	if err != nil {
		return err
	}

	if AppCnf.EnableDelFileBackup {
		if AppCnf.DelFileBackupDuration == 0 {
			AppCnf.DelFileBackupDuration = time.Hour * 72
		}

		if AppCnf.DelFileBackupPath == "" {
			AppCnf.DelFileBackupPath = path.Join(AppCnf.Path, "del_backup")
		}

		err := os.MkdirAll(AppCnf.DelFileBackupPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
