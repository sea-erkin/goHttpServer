package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

var (
	print              = fmt.Println
	listenPortFlag     = flag.String("p", "", "-p Port to listen on. Kinda optional, will use 80 if not provided")
	logFileFlag        = flag.String("l", "", "(optional) -l Log file to write access logs")
	logJSON            = flag.Bool("j", false, "(optional) -j Saves log results as JSON. Requires logfile to be provided")
	redirectHttpsFlag  = flag.Bool("r", false, "(optional) -r Redirect using port 80 to port 443")
	serveDirectoryFlag = flag.String("d", "", "(optional) -d Path to directory to serve")
	certChainPathFlag  = flag.String("c", "", "(optional) -c Path to cert chain")
	certPrivKeyFlag    = flag.String("k", "", "(optional) -k Path to cert private key")
	isTLS              = false
	logFileMutex	   = sync.Mutex{}
)

// Need to add an interceptor for the respoonse as well to log server responses.

func main() {

	checkFlags()

	http.Handle("/", logHandler(http.FileServer(http.Dir(*serveDirectoryFlag))))

	if isTLS {
		if *redirectHttpsFlag {
			go http.ListenAndServe(":80", logHandler(http.HandlerFunc(redirectHttpsHandler)))
		}
		log.Fatal(http.ListenAndServeTLS(":"+*listenPortFlag, *certChainPathFlag, *certPrivKeyFlag, nil))
	} else {
		log.Fatal(http.ListenAndServe(":"+*listenPortFlag, nil))
	}
}

func logHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLog := RequestLog{
			RemoteAddr: r.RemoteAddr,
			URL:        r.URL.String(),
			UserAgent:  r.UserAgent(),
			Referer:    r.Referer(),
			Method:     r.Method,
			RequestURI: r.RequestURI,
			Protocol:   r.Proto,
		}
		err := writeLog(requestLog)
		if err != nil {
			log.Fatal(err)
		}
		if *logFileFlag == "" {
			log.Printf("%s %s %s %s %s %s %s", requestLog.RemoteAddr, requestLog.URL, requestLog.UserAgent, requestLog.Referer, requestLog.Method, requestLog.RequestURI, requestLog.Protocol)
		}
		handler.ServeHTTP(w, r)
	})
}

func redirectHttpsHandler(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}

func writeLog(requestLog RequestLog) error {
	if *logFileFlag != "" {

		// check if log file exists
		_, err := os.Stat(*logFileFlag)
		logFileExists := false

		// create file dir if not exists
		if err != nil {
			dir := filepath.Dir(*logFileFlag)
			err := os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			logFileExists = true
		}

		if *logJSON {
			err := writeLogFileJson(logFileExists, requestLog)
			if err != nil {
				return err
			}
		} else {
			err := writeLogTab(requestLog)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeLogFileJson(logFileExists bool, logEntry RequestLog) error {
	if logFileExists {
		logBytes, err := ioutil.ReadFile(*logFileFlag)
		print(len(logBytes))
		if err != nil {
			return err
		}
		var logSlice = []RequestLog{}
		err = json.Unmarshal(logBytes, &logSlice)
		if err != nil {
			return err
		}
		logSlice = append(logSlice, logEntry)
		logJSON, _ := json.Marshal(logSlice)
		logFileMutex.Lock()
		defer logFileMutex.Unlock()
		err = ioutil.WriteFile(*logFileFlag, logJSON, 0644)
		if err != nil {
			return err
		}
	} else {
		// write log slice json
		var logSlice = []RequestLog{logEntry}
		logJSON, err := json.Marshal(logSlice)
		if err != nil {
			return err
		}
		logFileMutex.Lock()
		defer logFileMutex.Unlock()
		err = ioutil.WriteFile(*logFileFlag, logJSON, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeLogTab(requestLog RequestLog) error {
	f, err := os.OpenFile(*logFileFlag, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(io.MultiWriter(os.Stderr, f))
	log.Printf("%s %s %s %s %s %s %s", requestLog.RemoteAddr, requestLog.URL, requestLog.UserAgent, requestLog.Referer, requestLog.Method, requestLog.RequestURI, requestLog.Protocol)
	defer f.Close()
	return nil
}

func checkFlags() error {

	flag.Parse()
	if *listenPortFlag == "" {
		print("[INFO] No listen port provided, setting listen port to 80")
		*listenPortFlag = "80"
	}

	if *certChainPathFlag != "" {
		_, err := os.Stat(*certChainPathFlag)
		if err != nil {
			return errors.New("[ERROR] Cert chain path invalid")
		}
	}

	if *certPrivKeyFlag != "" {
		_, err := os.Stat(*certPrivKeyFlag)
		if err != nil {
			return errors.New("[ERROR] Cert private key path ivalid")
		}
	}

	if *listenPortFlag == "443" && (*certChainPathFlag == "" || *certPrivKeyFlag == "") {
		return errors.New("[ERROR] Provided port 443 but no certificate!")
	}

	if *certChainPathFlag != "" && *certPrivKeyFlag != "" {
		isTLS = true
	}

	if *logJSON && *logFileFlag == "" {
		return errors.New("[ERROR] Specified logging as JSON but did not provide log file path")
	}

	return nil
}

type RequestLog struct {
	RemoteAddr string
	URL        string
	UserAgent  string
	Referer    string
	Method     string
	RequestURI string
	Protocol   string
}
