package godaemon


import (
	"os"
	"net"
	"log"
	"errors"
	"fmt"
	"time"
	"strconv"
)

const (
	//environments
	SD_NOTIFY_SOCKET = "NOTIFY_SOCKET"
	SD_WATCHDGO_USEC = "WATCHDOG_USEC"
	// signals.
	DEAMON_READY = "READY=1"
	DEAMON_RELOADING = "RELOADING=1"
	DEAMON_STOPPING = "STOPPING=1"
	DEAMON_STATUS = "STATUS=%s"
	DEAMON_ERRNO = "ERRNO=%s"
	DEAMON_BUSERROR = "BUSERROR=%s"
	DEAMON_MAINPID = "MAINPID=%v"
	DEAMON_WATCHDOG = "WATCHDOG=1"
	DEAMON_FDSTORE = "FDSTORE=1"
	DEAMON_DFNAME = "FSNAME=%s"
	DEAMON_WATCHDOG_USEC = "WATCHDOG_USEC=%v"
)

//var notifySocketAddr string
var notifySocket net.Conn
var watchdogUpdateDuration time.Duration = 5 * time.Second

func init() {
	var err error
	notifySocketAddr := os.Getenv(SD_NOTIFY_SOCKET)
	dur := os.Getenv(SD_WATCHDGO_USEC)
	if len(dur) > 0 {
		if micsec, err := strconv.Atoi(dur); err == nil {
			watchdogUpdateDuration = time.Duration((micsec / 2)) * time.Microsecond
		}
	}

	if len(notifySocketAddr) > 0 {
		notifySocket, err = net.Dial("unixgram", notifySocketAddr)
		if err != nil {
			log.Printf("dial notify socket[%s] error:%v", notifySocketAddr, err)
		}
	}
}

func checkSocket() error {
	if notifySocket == nil {
		return errors.New("no notify socket")
	}
	return nil
}

func DeamonReady() (error) {
	return SD_Notify(DEAMON_READY)
}
func DeamonReload() (error) {
	return SD_Notify(DEAMON_RELOADING)
}
func DeamonStatus(status string) error {
	return SD_Notify(fmt.Sprintf(DEAMON_STATUS, status))
}

func DeamonPID() error {
	return SD_Notify(fmt.Sprintf(DEAMON_MAINPID, os.Getpid()))
}

func FeedWatchDog() error {
	return SD_Notify(DEAMON_WATCHDOG)
}

//KeepDeamonAlive
func KeepDeamonAlive() {
	if err := checkSocket(); err != nil {
		return
	}
	go func() {
		tmr := time.NewTicker(watchdogUpdateDuration)
		defer tmr.Stop()
		for {
			if _, ok := <-tmr.C; ok {
				FeedWatchDog()
			} else {
				break
			}

		}
	}()
}

func SD_Notify(msg string) (error) {
	if err := checkSocket(); err != nil {
		return err
	}
	if _, err := notifySocket.Write([]byte(msg)); err != nil {
		return err
	}
	return nil
}