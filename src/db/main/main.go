package main

import (
	"db"

	"bufio"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
)

func readUrls(dbase *db.Db, file string) error {
	h, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		logger.Error("Failed to open file: ", file, " error: ", err)
		return err
	}
	defer h.Close()
	scanner := bufio.NewScanner(h)
	for scanner.Scan() {
		line := scanner.Text()
		logger.Debug("Read URL: ", line)
		tokens := strings.Fields(line)
		if len(tokens) > 1 {
			logger.Debug("Adding URL: ", tokens[0], " type: ", tokens[1])
			dbase.Add(tokens[0], tokens[1])
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("Failed to scan file: ", err)
		return err
	}
	return nil
}

var (
	logger *logrus.Logger
)

func main() {
	host := flag.String("host", "127.0.0.1", "Server address.")
	port := flag.Int("port", 8888, "Server port.")
	prefix := flag.String("prefix", "/urlinfo/1/", "Document root.")
	fwdb := flag.String("fwdb", "", "Text file containing blacklisted URLs.")
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	logger = logrus.StandardLogger()

	dbase := db.New(*prefix)
	if *fwdb != "" {
		if err := readUrls(dbase, *fwdb); err != nil {
			return
		}
	}
	http.HandleFunc("/", dbase.HandleHttpGet)
	listener := fmt.Sprintf("%s:%d", *host, *port)
	logger.Error(http.ListenAndServe(listener, nil))
}
