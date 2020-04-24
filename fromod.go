// Copyright 2020 Kaur Kuut
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// The timezone of the video timestamp.
const tzLoc = "Europe/Tallinn"

// The location of the ffmpeg binary
const ffmpeg = `D:\Apps\ffmpeg\bin\ffmpeg.exe`

// The directory that contains the MOI/MOD files. Will also be used for output.
const workDir = `D:\Unsorted\THX Valuables\SD_VIDEO\`

var timeLocation *time.Location

func main() {
	loc, err := time.LoadLocation(tzLoc)
	if err != nil {
		fmt.Printf("Failed to load time location %v: %v\n", tzLoc, err)
		return
	}
	timeLocation = loc

	dealWithFiles(workDir)
}

func dealWithFiles(dirName string) {
	// Get the file list for this directory
	fileInfos := getFileList(dirName)

	for j := range fileInfos {
		name := fileInfos[j].Name()
		fullName := filepath.Join(dirName, name)
		isDir := fileInfos[j].IsDir()

		if isDir {
			dealWithFiles(fullName)
		} else {
			if strings.HasSuffix(name, ".MOI") {
				processFile(fullName)
			}
		}
	}
}

func getFileList(dirName string) []os.FileInfo {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		fmt.Printf("ReadDir failed: %v\n", err)
		panic("")
	}
	return files
}

func processFile(filePath string) {
	t := getDateFromMOI(filePath)

	modPath := filePath[:len(filePath)-3] + "MOD"
	mpgPath := filepath.Join(filepath.Dir(modPath), fmt.Sprintf("%v.mpg", t.Format("2006-01-02--15-04-05")))

	cmd := exec.Command(ffmpeg, "-y", "-i", modPath, "-vcodec", "copy", "-acodec", "copy", mpgPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Failed to run ffmpeg on %v because: %v\nHere's the output:\n\n%v\n", modPath, err, string(out))
	} else {
		fmt.Printf("%v > %v\n", modPath, filepath.Base(mpgPath))
	}
	if err := os.Chtimes(mpgPath, t, t); err != nil {
		fmt.Printf("Failed to set file time for %v", filepath.Base((mpgPath)))
	}
}

func getDateFromMOI(filePath string) time.Time {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read MOI file: %v", err))
	}

	if len(data) < 14 {
		panic(fmt.Sprintf("Invalid MOI file length %v", len(data)))
	}

	year := int(data[6])<<8 | int(data[7])
	month := int(data[8])
	day := int(data[9])
	hour := int(data[10])
	min := int(data[11])
	millisec := int(data[12])<<8 | int(data[13])

	return time.Date(year, time.Month(month), day, hour, min, 0, millisec*1000000, timeLocation)
}
