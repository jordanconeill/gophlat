package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	cp "github.com/nmrshll/go-cp"
)

var SKIPLOG string = "./skip.log"
var PHLATLOG string = "./phlat.log"
var ERRLOG string = "./errors.log"

func main() {
	if len(os.Args[1:]) > 2 {
		fmt.Printf("\nToo many arguments provided.\n\nUsage: gophlat <target directory> <output directory>\n\n")
	}

	tgtdir := os.Args[1]
	// fmt.Printf("TEST: Tgtdir: %s\n", tgtdir) //TEST
	outdir := os.Args[2]
	// fmt.Printf("TEST: Outdir: %s\n", outdir) //TEST

	// Check if outdir exists
	outDirInfo, err := os.Stat(outdir)
	if os.IsNotExist(err) {
		// Create directory
		err := os.MkdirAll(outdir, os.ModePerm)
		check(err)
	} else if !outDirInfo.IsDir() {
		// Dir provided is a file
		fmt.Printf("The <output directory> argument must be a directory! (%s)\n", outdir)
		os.Exit(1)
	} else {
		// Check if dir is empty
		empty, _ := isEmpty(outdir)
		if !empty {
			var consent string
			fmt.Print("The output directory provided is not empty. Some files may be overwritten. Do you wish to continue? (y/N): ") //TODO: file.txt and file(1).txt may both be present in dst. If file.txt is being copied, file(1).txt will be overwritten. Solution: if "y", add files and hashes in dst to collisionMap.
			fmt.Scanln(&consent)
			if strings.ToLower(consent) != "y" && strings.ToLower(consent) != "yes" {
				fmt.Println("Program terminated by user.")
				// fmt.Printf("TEST: Consent to lower: %s\n", strings.ToLower(consent)) //TEST
				os.Exit(0)
			}
		}
	}

	// Delete old log files
	_ = os.Remove(SKIPLOG)
	_ = os.Remove(PHLATLOG)
	_ = os.Remove(ERRLOG)

	// Initialize log files
	StampLogs()

	// Get unique (by sha256 filehash) files for flattening
	phlats, err := getPhlats(tgtdir)
	check(err)
	// PrintHashMap(phlats) //TEST

	// Copy files to specified output directory (outdir)
	collisionMap := make(map[string]int)
	for _, v := range phlats {
		// fmt.Printf("TEST: file to copy: %s\n", v) //TEST
		err := CopyFile(v, outdir, collisionMap)
		check(err)
	}

	StampLogs()
}

func check(e error) {
	if e != nil {
		log.Fatalf("%s\n", e)
	}
}

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func getPhlats(tgtdir string) (map[string]string, error) {
	phlats := make(map[string]string)

	err := filepath.Walk(tgtdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if info.IsDir() {
			// fmt.Printf("TEST: %s is a dir\n", path) //TEST
			err := logSkip(path, "Object is a directory.")
			if err != nil {
				err = logErr(err)
				if err != nil {
					fmt.Printf("Warning: Error logging to ERRLOG: %e\n", err)
				}
			}
			return nil
		} else {
			// Get file hash
			hash := HashFile(path)
			// Check if hash is not duplicated
			if _, keyExists := phlats[hash]; !keyExists {
				// Add file to list
				phlats[hash] = path
			} else {
				err := logSkip(path, "Duplicate file hash.")
				if err != nil {
					err = logErr(err)
					if err != nil {
						fmt.Printf("Warning: Error logging to ERRLOG: %e\n", err)
					}
				}
			}
		}
		return nil
	})
	return phlats, err
}

func HashFile(file string) string {
	input, err := os.Open(file)
	check(err)
	defer input.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, input); err != nil {
		log.Fatal(err)
	}
	sum := hash.Sum(nil)
	return string(sum[:])
}

// func PrintHashMap(m map[string]string) { //TEST
// 	fmt.Println("\nTEST:\nFileName\tFileHash")
// 	for k, v := range m {
// 		fmt.Printf("%s\t%x\n", v, k)
// 	}
// }

func CopyFile(src string, dst string, collisionMap map[string]int) error {
	dstpath := filepath.Join(dst, filepath.Base(src))
	// fmt.Printf("TEST: dstpath: %s\nChecking if file exists in destination...\n", dstpath) //TEST
	// fmt.Printf("TEST: filepath.Base(src): %s\n", filepath.Base(src))
	_, err := os.Stat(dstpath)
	if !os.IsNotExist(err) {
		// Filename collision, check if collision was previously recorded
		// fmt.Println("TEST: Collision!") //TEST
		if _, keyExists := collisionMap[filepath.Base(src)]; !keyExists {
			// Add file to list
			collisionMap[filepath.Base(src)] = 1
			// fmt.Printf("TEST: New collisionMap entry: %d\n", collisionMap[filepath.Base(src)]) //TEST
		} else {
			collisionMap[filepath.Base(src)] = collisionMap[filepath.Base(src)] + 1
			// fmt.Printf("TEST: collisionMap updated: %d\n", collisionMap[filepath.Base(src)]) //TEST
		}
		// Add (1) or (2) or (n) to destination filename, e.g. file(1).txt
		dstpath = dstpath[:len(dstpath)-len(filepath.Ext(dstpath))] + "(" + fmt.Sprint(collisionMap[filepath.Base(src)]) + ")" + filepath.Ext(dstpath)
		// fmt.Printf("TEST: New dstpath: %s\n", dstpath) //TEST
	}
	err = cp.CopyFile(src, dstpath)
	if err != nil {
		return err
	}
	err = logPhlat(filepath.Base(dstpath))
	if err != nil {
		err = logErr(err)
		if err != nil {
			fmt.Printf("Warning: Error logging to ERRLOG: %e\n", err)
		}
	}
	return nil
}

func StampLogs() {
	// SKIPLOG
	skipFile, err := os.OpenFile(SKIPLOG, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Printf("Warning: Error initializing SKIPLOG: %e", err)
	}
	defer skipFile.Close()
	_, err = skipFile.WriteString(fmt.Sprintf("%s\n", time.Now()))
	if err != nil {
		fmt.Printf("Warning: (initialization) Error writing to SKIPLOG: %e", err)
	}

	// PHLATLOG
	phlatFile, err := os.OpenFile(PHLATLOG, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Printf("Warning: Error initializing PHLATLOG: %e", err)
	}
	defer phlatFile.Close()
	_, err = phlatFile.WriteString(fmt.Sprintf("%s\n", time.Now()))
	if err != nil {
		fmt.Printf("Warning: (initialization) Error writing to PHLATLOG: %e", err)
	}

	// ERRLOG
	errFile, err := os.OpenFile(ERRLOG, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		fmt.Printf("Warning: Error initializing ERRLOG: %e", err)
	}
	defer errFile.Close()
	_, err = errFile.WriteString(fmt.Sprintf("%s\n", time.Now()))
	if err != nil {
		fmt.Printf("Warning: (initialization) Error writing to ERRLOG: %e", err)
	}
}

func logSkip(path string, reason string) error {
	logfile, err := os.OpenFile(SKIPLOG, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer logfile.Close()

	_, err = logfile.WriteString(fmt.Sprintf("File \"%s\" skipped. Reason: %s\n", path, reason))
	if err != nil {
		return err
	}

	return nil
}

func logPhlat(path string) error {
	logfile, err := os.OpenFile(PHLATLOG, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer logfile.Close()

	_, err = logfile.WriteString(fmt.Sprintf("%s\n", path))
	if err != nil {
		return err
	}

	return nil
}

func logErr(e error) error {
	logfile, err := os.OpenFile(ERRLOG, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer logfile.Close()

	_, err = logfile.WriteString(fmt.Sprintf("%s\n", e))
	if err != nil {
		return err
	}

	return nil
}
