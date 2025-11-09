package internal

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DeleteFileReq struct {
	FilePath *string `json:"file_path,omitempty" xml:"file_path,omitempty" form:"file_path,omitempty"`
}

// HandleDeleteFile will require API-KEY & API-SECRET as header value
func HandleDeleteFile(c *fiber.Ctx) error {
	apiKey := c.Get("API-KEY")
	secret := c.Get("API-SECRET")

	if apiKey == "" || secret == "" {
		return sendResponse(c, fiber.StatusUnauthorized, false, "Auth header information are missing")
	}

	if apiKey != AppCnf.ApiKey || secret != AppCnf.ApiSecret {
		return sendResponse(c, fiber.StatusUnauthorized, false, "Auth header information didn't match")
	}

	req := new(DeleteFileReq)
	err := c.BodyParser(req)
	if err != nil {
		return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
	}

	if req.FilePath == nil {
		return sendResponse(c, fiber.StatusNotAcceptable, false, "file_path value require")
	}

	file := fmt.Sprintf("%s/%s", AppCnf.Path, *req.FilePath)
	f, err := os.Stat(file)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			return sendResponse(c, fiber.StatusNotFound, false, *req.FilePath+" does not exist")
		} else {
			return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
		}
	}

	if f.IsDir() {
		return sendResponse(c, fiber.StatusNotAcceptable, false, *req.FilePath+" is a directory, not allow to delete directory")
	}

	if AppCnf.EnableDelFileBackup {
		// first with the video file
		toFile := path.Join(AppCnf.DelFileBackupPath, f.Name())
		err := os.Rename(file, toFile)
		if err != nil {
			log.Println(err)
			return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
		}

		// otherwise during cleanup will be hard to detect
		newTime := time.Now()
		if err := os.Chtimes(toFile, newTime, newTime); err != nil {
			log.Println("Failed to update file modification time:", err)
		}
	} else {
		err = os.Remove(file)
		if err != nil {
			return sendResponse(c, fiber.StatusNotAcceptable, false, err.Error())
		}
	}

	if AppCnf.Compress {
		// silently delete compressed file
		_ = os.Remove(file + ".fiber.gz")
	}

	if AppCnf.DeleteEmptyDir {
		dir := strings.Replace(file, "/"+f.Name(), "", 1)
		if dir != AppCnf.Path {
			if empty, err := IsDirEmpty(dir); err == nil && empty {
				_ = os.Remove(dir)
			}
		}
	}

	return sendResponse(c, fiber.StatusOK, true, "file deleted")
}
