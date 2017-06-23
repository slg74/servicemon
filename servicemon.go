package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

type email struct {
	From                 string
	NoReplyAcct          string
	To                   string
	ToAcct               string
	Subject              string
	TxtHTMLBody          string
	ProftpdIsDownMsg     string
	ProftpdIsDownBody    string
	ProftpdRestartedMsg  string
	ProftpdRestartedBody string
	SMTPServer           string
	SMTPPort             int
}

var serviceNames []string = []string{
	"r1ctl",
	"r1rm",
	"cdpserver",
	"proftpd",
}

func listServices() {
	for i := range serviceNames {
		fmt.Println(serviceNames[i])
	}
}


func getNoreplyPassword() string {
	file, err := ioutil.ReadFile("/home/continuum/bin/.noreplypw")
	if err != nil {
		fmt.Println(err.Error())
	}
	pw := string(file)
	return pw
}

func proftpdIsDown() bool {
	cmd := "ps -ef|grep [p]roftpd >/dev/null 2>&1; echo $?"

	status, err := exec.Command("bash","-c",cmd).Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: %s\n", cmd)
	}

	i, _ := strconv.Atoi(strings.Trim(string(status), "\n"))
	if i == 0 {
		return true
	}

	return false
}

func proftpdRestarted() bool {
	cmd := "service proftpd restart"
	status, err := exec.Command("bash","-c",cmd).Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: %s\n", cmd)
	}

	i, _ := strconv.Atoi(strings.Trim(string(status), "\n"))
	if i != 0 {
		fmt.Printf("Unable to restart proftpd: ", err.Error())
		return false
	}
	return true
}

func serviceIsDown(serviceName string) bool {
	cmd := "service " + serviceName + " status"
	status, err := exec.Command("bash","-c",cmd).Output()
	if err != nil {
		fmt.Sprintf("Failed to execute command: %s\n", cmd)
	}
	i, _ := strconv.Atoi(strings.Trim(string(status), "\n"))
	if i == 0 {
		return true
	}
	return false
}

func main() {
	
	listServices()

	e := email{}

	e.From		= "From"
	e.NoReplyAcct   = "noreply@r1soft.com"
	e.To		= "To"
	//e.ToAcct        = "c247devops@r1soft.com"
	e.ToAcct        = "scott.gillespie@r1soft.com"
	e.TxtHTMLBody   = "text/html"
	e.Subject       = "Subject"
	e.SMTPServer    = "smtp.office365.com"
	e.SMTPPort      = 587

	m := gomail.NewMessage()

	m.SetHeader(e.From, e.NoReplyAcct)
	m.SetHeader(e.To, e.ToAcct)

	hostname, _ := os.Hostname()
	pw := strings.Trim(getNoreplyPassword(), "\n")

	if (serviceIsDown("r1ctl")) {
		fmt.Printf("r1ctl process is down on %s\n", hostname)
	}

	if (proftpdIsDown()) {

		fmt.Printf("proftpd is down on %s.\n", hostname)

		e.ProftpdIsDownMsg    = "PROFTPD DOWN ON " + hostname
		e.ProftpdIsDownBody   = "proftpd server is down on host: " + hostname

		m.SetHeader(e.Subject, e.ProftpdIsDownMsg)
		m.SetBody(e.TxtHTMLBody, e.ProftpdIsDownBody)

		d := gomail.NewDialer(e.SMTPServer, e.SMTPPort, e.NoReplyAcct, pw)

		err := d.DialAndSend(m)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("restarting proftpd on %s.\n", hostname)


		if (proftpdRestarted()) {

			fmt.Printf("proftpd restarted successfully on %s.\n", hostname)
			e.ProftpdRestartedMsg   = "PROFTPD RESTARTED ON " + hostname
			e.ProftpdRestartedBody  = "proftpd restarted on host: " + hostname
			m.SetHeader(e.Subject, e.ProftpdRestartedMsg)
			m.SetBody(e.TxtHTMLBody, e.ProftpdRestartedBody)
			d := gomail.NewDialer(e.SMTPServer, e.SMTPPort, e.NoReplyAcct, pw)
			err := d.DialAndSend(m)
			if err != nil {
				log.Fatal(err)
			}

		} else {

			fmt.Printf("unable to restart proftpd on %s. Please investigate.\n", hostname)
			e.ProftpdRestartedMsg   = "PROFTPD DID NOT RESTART ON " + hostname
			e.ProftpdRestartedBody  = "proftpd failed to restart on host: " + hostname
			m.SetHeader(e.Subject, e.ProftpdRestartedMsg)
			m.SetBody(e.TxtHTMLBody, e.ProftpdRestartedBody)
			d := gomail.NewDialer(e.SMTPServer, e.SMTPPort, e.NoReplyAcct, pw)
			err := d.DialAndSend(m)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
