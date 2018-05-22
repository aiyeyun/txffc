package logger

import (
	"os"
	"time"
	"fmt"
	"log"
	"sync"
)

func init()  {
	gopath := os.Getenv("GOPATH")
	logpath := gopath + "/src/txffc/log/"
	logspath := gopath + "/src/txffc/logs/"

	f, err := os.Open(logpath)
	defer f.Close()
	if err != nil {
		os.Mkdir(logpath, 0777)
	}

	logsf, logserr := os.Open(logspath)
	defer logsf.Close()
	if logserr != nil {
		os.Mkdir(logspath, 0777)
	}
}

func Log(contents string)  {
	m := new(sync.Mutex)
	m.Lock()

	gopath := os.Getenv("GOPATH")
	logpath := gopath + "/src/txffc/log/"

	logfile := logpath + time.Now().Format("2006-01-02")

	f, err := os.Open(logfile)
	if err != nil {
		os.Create(logfile)
	}

	logf ,err := os.OpenFile(logfile, os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println(err)
	}
	log_ger := log.New(logf, "\r\n", log.Ldate|log.Ltime)
	log_ger.Println("进程ID:", os.Getpid(), contents)
	log.Println("进程ID:", os.Getpid(), contents)

	defer func() {
		m.Unlock()
		f.Close()
		logf.Close()
	}()
}

func Logs(contents string, filename string)  {
	m := new(sync.Mutex)
	m.Lock()

	gopath := os.Getenv("GOPATH")
	logpath := gopath + "/src/txffc/logs/"

	logfile := logpath + filename

	f, err := os.Open(logfile)
	if err != nil {
		os.Create(logfile)
	}

	logf ,err := os.OpenFile(logfile, os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println(err)
	}
	log_ger := log.New(logf, "\r\n", log.Ldate|log.Ltime)
	log_ger.Println("进程ID:", os.Getpid(), contents)
	//log.Println("进程ID:", os.Getpid(), contents)

	defer func() {
		m.Unlock()
		f.Close()
		logf.Close()
	}()
}