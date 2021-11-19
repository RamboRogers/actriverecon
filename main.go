package main

/*

ActiveRecon

Author: matt@matthewrogers.org
Date: 06/01/2021

License: GNU GPLv3

*/

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//Nmaprun scan format
type Nmaprun struct {
	XMLName          xml.Name `xml:"nmaprun"`
	Text             string   `xml:",chardata"`
	Scanner          string   `xml:"scanner,attr"`
	Start            string   `xml:"start,attr"`
	Version          string   `xml:"version,attr"`
	Xmloutputversion string   `xml:"xmloutputversion,attr"`
	Scaninfo         struct {
		Text     string `xml:",chardata"`
		Type     string `xml:"type,attr"`
		Protocol string `xml:"protocol,attr"`
	} `xml:"scaninfo"`
	Host []struct {
		Text    string `xml:",chardata"`
		Endtime string `xml:"endtime,attr"`
		Address struct {
			Text     string `xml:",chardata"`
			Addr     string `xml:"addr,attr"`
			Addrtype string `xml:"addrtype,attr"`
		} `xml:"address"`
		Ports struct {
			Text string `xml:",chardata"`
			Port struct {
				Text     string `xml:",chardata"`
				Protocol string `xml:"protocol,attr"`
				Portid   string `xml:"portid,attr"`
				State    struct {
					Text      string `xml:",chardata"`
					State     string `xml:"state,attr"`
					Reason    string `xml:"reason,attr"`
					ReasonTtl string `xml:"reason_ttl,attr"`
				} `xml:"state"`
			} `xml:"port"`
		} `xml:"ports"`
	} `xml:"host"`
	Runstats struct {
		Text     string `xml:",chardata"`
		Finished struct {
			Text    string `xml:",chardata"`
			Time    string `xml:"time,attr"`
			Timestr string `xml:"timestr,attr"`
			Elapsed string `xml:"elapsed,attr"`
		} `xml:"finished"`
		Hosts struct {
			Text  string `xml:",chardata"`
			Up    string `xml:"up,attr"`
			Down  string `xml:"down,attr"`
			Total string `xml:"total,attr"`
		} `xml:"hosts"`
	} `xml:"runstats"`
}

//Scan database format
type Scan struct {
	gorm.Model
	Host     string
	Addr     string
	Protocol string
	Port     string
	Service  string
}

var defaultScanConfig = `#Rate over 1000 can cause major issues for a docker host
#Rate over 20000 can cause issues over a 1Gbit NIC
rate =   1000.00
randomize-hosts = true
seed = 5804244860458012382
shard = 1/1
# ADAPTER SETTINGS
adapter = 
adapter-ip = 0.0.0.0
adapter-mac = 00:00:00:00:00:00
router-mac = 00:00:00:00:00:00
# OUTPUT/REPORTING SETTINGS
output-format = xml
show = open,,
output-filename = scan.xml
rotate = 0
rotate-dir = .
rotate-offset = 0
rotate-filesize = 0
pcap = 
# TARGET SELECTION (IP, PORTS, EXCLUDES)
retries = 0
ports = 0-10000
range = 192.168.1.0/24,192.168.2.0/24,192.168.5.0/24,192.168.10.0/24,192.168.3.0/24

capture = cert
#nocapture = html
#nocapture = heartbleed
#nocapture = ticketbleed

min-packet = 60`

//Global Database
var db *gorm.DB

var webPageTopTemplate = `
<!doctype html>
<html lang="en">
<head>
<!-- Required meta tags -->
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<!-- Latest compiled and minified CSS -->
<link rel="stylesheet" href="/css/bootstrap.min.css">
<!-- Latest compiled JavaScript -->
<script src="/js/jquery.min.js"></script>
<script src="/js/bootstrap.bundle.min.js"></script>

<script>
const sleep = (milliseconds) => {
	return new Promise(resolve => setTimeout(resolve, milliseconds))
  }

async function getText(url){
	document.getElementById("textOutput").innerHTML = "<div class=\"spinner-border text-success\"></div> Running API Call to URL:" + url + "<br><img src=\"static/hacking.gif\">";
	let myObject = await fetch(url);
	let myText = await myObject.text();
	sleep(500).then(() => {
		document.getElementById("textOutput").innerHTML = myText;
	})	
}` +
	"\n" +
	"function resetConfig() { \n" +
	"document.getElementById(\"configData\").value = `" + defaultScanConfig + "`;" +
	"\n} \n" +
	`</script>

<title>ActiveRecon</title>
</head>
<body>
<div class="jumbotron text-center">
  <h1>ActiveRecon</h1>
  <p>The amazing simple to use recon tool.</p>
  <button type="button" class="btn btn-info" id="clean" onclick="getText('/editConfig')" data-toggle="tooltip" title="✍️ This edits the scan config file used by masscan." >Edit Config</button> 
  <button type="button" class="btn btn-warning" id="performScan" onclick="getText('/performScan')" data-toggle="tooltip" title="🏃‍♂️ This runs the network scan.  Depending on subnet size, this can take a long time, or a short time. 🚀" >Perform a Scan</button>
  <button type="button" class="btn btn-primary" id="readScan" onclick="getText('/readScan')" data-toggle="tooltip" title="📥 This reads the latest scan output, should take a few seconds." >Read Latest Scan</button> 
  <button type="button" class="btn btn-success" id="getScreenShots" onclick="getText('/getScreenShots')" data-toggle="tooltip" title="⏳ This can take around 20 seconds per host. So get a cup of coffee." >Get Screen Shots</button> | 
  <button type="button" class="btn btn-danger" id="clean" onclick="getText('/clean')" data-toggle="tooltip" title="🧨 WARNING: This purges all data and wipes the database!" >Wipe Database</button>
  <P>
  <div id="textOutput" class="container-fluid border bg-light"></div>
</div>

<div class="container">

`

//getText() needs to open some sort of a DIV or IFRAME with the content inside of it.

var webPageBottomTemplate = `
</div>
<script>
$(document).ready(function(){
  $('[data-toggle="tooltip"]').tooltip(); 
});
</script>
</body>
</html>
`

//PortURL for building URLs
var knownPortURLs = map[string]string{
	"80":   "http://",
	"8181": "http://",
	"9090": "http://",
	"81":   "http://",
	"7777": "http://",
	"3000": "http://",
	"8002": "http://",
	"8001": "http://",
	"8080": "http://",
	"8096": "http://",
	"443":  "https://",
	"22":   "ssh://",
	"3389": "rdp://",
	"5900": "vnc://",
	"445":  "cifs://",
}

func main() {

	version := "v0.001"
	fmt.Println("ActiveRecon " + version + " - Started " + time.Now().Format(time.RFC850))

	//open db
	var err error
	db, err = gorm.Open(sqlite.Open("activedata.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Scan{})

	listenAddress := ":9009"

	fs := http.FileServer(http.Dir("visual/"))
	http.Handle("/visual/", http.StripPrefix("/visual/", fs))

	fs1 := http.FileServer(http.Dir("css/"))
	http.Handle("/css/", http.StripPrefix("/css/", fs1))

	fs2 := http.FileServer(http.Dir("js/"))
	http.Handle("/js/", http.StripPrefix("/js/", fs2))

	fs3 := http.FileServer(http.Dir("screenshots/"))
	http.Handle("/screenshots/", http.StripPrefix("/screenshots/", fs3))

	fs4 := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs4))

	http.HandleFunc("/clean", performClean)
	http.HandleFunc("/readScan", readScan)
	http.HandleFunc("/performScan", performScan)
	http.HandleFunc("/getScreenShots", getScreenShots)
	http.HandleFunc("/editConfig", editConfig)
	http.HandleFunc("/writeConfig", writeConfig)
	http.HandleFunc("/", presentMenu)

	log.Println("Server is up....", listenAddress)

	//Server is parallel
	log.Fatal(http.ListenAndServe(listenAddress, nil))

}

func editConfig(w http.ResponseWriter, r *http.Request) {
	ip, _ := getIP(r)
	log.Println("editConfig Called by " + ip)
	data, _ := ioutil.ReadFile("settings.scan")
	fmt.Fprintf(w, "<h1>Editing %s</h1>"+
		"<form action=\"/writeConfig\" method=\"POST\">"+
		"<textarea id=\"configData\" name=\"body\" rows=\"35\" cols=\"80\">%s</textarea><br>"+
		"<button type=\"button\" class=\"btn btn-danger\" id=\"default\" onclick=\"resetConfig()\">Reset Config</button> | "+
		//"<input type=\"submit\" class\"btn btn-success\" value=\"Save\">"+
		"<button class=\"btn btn-success\" type=\"submit\">Save</button>"+
		"</form>",
		"settings.scan", data)
}

func writeConfig(w http.ResponseWriter, r *http.Request) {
	ip, _ := getIP(r)
	log.Println("writeConfig Called by " + ip)
	log.Println("Writing Config")
	filename := "settings.scan"
	body := r.FormValue("body")
	//log.Println("Body: " + body)
	ioutil.WriteFile(filename, []byte(body), 0600)
	fmt.Fprintf(w, "<html><body>Saved.")
	fmt.Fprintf(w, `<script>
	  location.replace("/")	
	</script>`)
}

//perform scan -- can't seem to find it.....
func performScan(w http.ResponseWriter, req *http.Request) {
	ip, _ := getIP(req)
	log.Println("performScan Called by " + ip)
	fmt.Fprintf(w, "Scan is starting.....\n")
	out, err := exec.Command("masscan", "-c", "settings.scan").Output()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Scan ouput: %v\n", out)
	fmt.Fprintf(w, "Scan Complete.")
}

func performClean(w http.ResponseWriter, req *http.Request) {
	ip, _ := getIP(req)
	log.Println("performClean Called by " + ip)

	//fmt.Fprintf(w, webPageTopTemplate)
	fmt.Fprintf(w, "<h1>Cleaning...</h1>")

	db.Table("scans").Where("1 = 1").Delete("%")

	fmt.Fprintf(w, "Done. ")
	fmt.Fprintf(w, "\n<a href=\"/\">Okay</a>\n")

	//fmt.Fprintf(w, webPageBottomTemplate)
}

//read scan
func readScan(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "<html>Scan Read in Progress....")
	ip, _ := getIP(req)
	log.Println("readScan Called by " + ip)
	//run a scan
	//performScan()

	// Open our xmlFile
	xmlFile, err := os.Open("scan.xml")
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Println(err)
	}
	log.Println("Successfully Opened scan.xml")
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	// we initialize our array
	var nmaprun Nmaprun

	//load the scan into this data structure
	xml.Unmarshal(byteValue, &nmaprun)

	//header info
	//fmt.Fprintf(w, "Scanner: "+nmaprun.Scanner+"\n")
	//fmt.Fprintf(w, "Scan Type: "+nmaprun.Scaninfo.Type+"\n")
	//fmt.Fprintf(w, "Hosts Up: "+nmaprun.Runstats.Hosts.Up+"\n")

	for i := 0; i < len(nmaprun.Host); i++ {
		//Host := nmaprun.Host[i].Address.Addr
		//Addr := nmaprun.Host[i].Address.Addr
		//Protocol := nmaprun.Host[i].Ports.Port.Protocol
		//Portid := nmaprun.Host[i].Ports.Port.Portid
		//fmt.Fprintf(w, "host:"+Host+" addr:"+Addr+" protocol:"+Protocol+" port:"+Portid+"\n")
		db.Create(&Scan{Host: nmaprun.Host[i].Address.Addr, Addr: nmaprun.Host[i].Address.Addr, Protocol: nmaprun.Host[i].Ports.Port.Protocol, Port: nmaprun.Host[i].Ports.Port.Portid})
	}
	fmt.Fprintf(w, "Database Updated....")
	//read database
	var scan []Scan
	/*
		// list all identified hosts
		db.Distinct("host").Order("host").Find(&scan)
		for _, host := range scan {
			fmt.Fprintf(w, "DBOUTPUT ->"+host.Host+"\n")

			//query open ports for host
			db.Distinct("host", "protocol", "port").Where("host = ?", host.Host).Order("port").Find(&scan)

			for _, host := range scan {
				fmt.Fprintf(w, "DBOUTPUT PORT ->"+host.Port+"\n")
			}
		}
	*/
	//select all hosts distinct and disctint ports
	//query open ports for host

	/*
		{
					"name": "network",
					"children": [
					{
					"name": "192.168.1.1",
					"children": [
					{
						"name": "ports",
						"children": [
						{"name": "8080", "value": 8080},
						{"name": "8181", "value": 8181}
						]
						}
					},
					{
					"name": "192.168.1.2",
					"children": [
						{
						"name": "ports",
						"children": [
						{"name": "8080", "value": 8080},
						{"name": "8181", "value": 8181}
						]
						}
					}

				]}
	*/
	//output to D3Data JSON
	//start json object
	jsonOut, err := os.Create("output.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintf(jsonOut, `{
		"name": "network",
		"children": [ 
		`)

	//distinct host query
	db.Distinct("host").Order("host").Find(&scan)
	hostCount := len(scan)
	for hostIndex, host := range scan {

		//start host json object
		fmt.Fprintf(jsonOut, "{\"name\":\"%v\",", host.Host)
		fmt.Fprintf(jsonOut, "\"children\": [ {")

		//start port json object / repeat
		fmt.Fprintf(jsonOut, "\"name\":\"%v\",", "ports")
		fmt.Fprintf(jsonOut, "\"children\": [ ")

		//lets get the ports
		db.Distinct("host", "protocol", "port").Where("host = ?", host.Host).Order("port").Find(&scan)

		//foreach port
		for index, host := range scan {

			//we add in the service info here, for now we'll just use port 2x
			fmt.Fprintf(jsonOut, "{\"name\": \"%v\", \"value\": \"%v\"}", host.Port, host.Port)

			if index < (len(scan) - 1) {
				fmt.Fprintf(jsonOut, ",")
			}
		}
		//close port
		fmt.Fprintf(jsonOut, "]}")

		//close host children
		fmt.Fprintf(jsonOut, "]}")
		//fmt.Printf("hostindex %v < %v\n", hostIndex, (hostCount - 1))
		if hostIndex < (hostCount - 1) {
			fmt.Fprintf(jsonOut, ",")
		}
	}
	//close object
	fmt.Fprintf(jsonOut, `]}`)

	fmt.Fprintf(w, "JSON for Visualizations Output....")
	//copy to visualizations
	visualDBFiles := [...]string{
		"visual/tidy-tree/files/e65374209781891f37dea1e7a6e1c5e020a3009b8aedf113b4c80942018887a1176ad4945cf14444603ff91d3da371b3b0d72419fa8d2ee0f6e815732475d5de",
		"visual/radial-dendrogram/files/e65374209781891f37dea1e7a6e1c5e020a3009b8aedf113b4c80942018887a1176ad4945cf14444603ff91d3da371b3b0d72419fa8d2ee0f6e815732475d5de",
		"visual/indented-tree/files/e65374209781891f37dea1e7a6e1c5e020a3009b8aedf113b4c80942018887a1176ad4945cf14444603ff91d3da371b3b0d72419fa8d2ee0f6e815732475d5de"}

	for _, file := range visualDBFiles {
		original, err := os.Open("output.json")
		if err != nil {
			log.Fatal(err)
		}
		defer original.Close()

		new, err := os.Create(file)
		if err != nil {
			log.Fatal(err)
		}
		defer new.Close()
		//This will copy
		_, err2 := io.Copy(new, original)
		if err != nil {
			log.Fatal(err2)
		}

	}

	err = jsonOut.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintf(w, "Done.\n")
	fmt.Fprintf(w, "\n<a href=\"/\">Okay</a>\n")
}

//the menu of links
//this also prints the current database
func presentMenu(w http.ResponseWriter, req *http.Request) {
	ip, _ := getIP(req)
	log.Println("presentMenu Called by " + ip)

	fmt.Fprintf(w, webPageTopTemplate)

	fmt.Fprintf(w, "<h3>Views</h3>")
	fmt.Fprintf(w, "<ul>")
	fmt.Fprintf(w, "<li><a target=\"_blank\" href=\"/visual/tidy-tree/index.html\">Visual Tidy-Tree</a></li>")
	fmt.Fprintf(w, "<li><a target=\"_blank\" href=\"/visual/radial-dendrogram/index.html\">Visual Radial-Dendrogram</a></li>")
	fmt.Fprintf(w, "<li><a target=\"_blank\" href=\"/visual/indented-tree/index.html\">Visual Indented-Tree</a></li>")
	fmt.Fprintf(w, "</ul>")

	fmt.Fprintf(w, "<h3>Network</h3>")
	fmt.Fprintf(w, "<table class=\"table\">")
	fmt.Fprintf(w, `
	<thead class="table-dark"><tr><td>IP</td> <td>Ports</td></tr><thead>
	<tbody class="table-light">
	`)

	var scan []Scan
	// list all identified hosts
	db.Distinct("host").Order("host").Find(&scan)
	for _, host := range scan {
		fmt.Fprintf(w, "<tr><td><h3>"+host.Host+"</h3></td>\n")

		//query open ports for host
		db.Distinct("host", "protocol", "port").Where("host = ?", host.Host).Order("port").Find(&scan)
		fmt.Fprintf(w, "<td><table class=\"table table-bordered border-secondary table-secondary table-active\">")
		for _, host := range scan {

			//make hotlinks for known ports.
			value, ok := knownPortURLs[host.Port]

			// Make a Regex to say we only want letters and numbers
			reg, err := regexp.Compile("[^a-zA-Z0-9]+")
			if err != nil {
				log.Fatal(err)
			}
			processedString := reg.ReplaceAllString(value, "")
			image := ""
			url := host.Port
			if ok {

				//does the image exist?
				_, err := os.Stat(("screenshots/" + processedString + "-" + host.Host + "-" + host.Port + ".png"))
				if os.IsNotExist(err) {
					//log.Println("File does not exist.")
				} else {
					image = "<a href=\"screenshots/" + processedString + "-" + host.Host + "-" + host.Port + ".png\" target=\"_blank\"><img height=200 width=300 src=\"screenshots/" + processedString + "-" + host.Host + "-" + host.Port + ".png\"></a>"
				}

				url = "<a target=\"_blank\" href=\"" + value + host.Host + ":" + host.Port + "\">" + host.Port + " (" + processedString + ")</td><td>" + image + "&nbsp;</a>"
			} else {
				url = url + "</td><td>&nbsp;</a>"
			}

			fmt.Fprintf(w, "<tr><td width=105>"+url+"\n")
		}
		fmt.Fprintf(w, "</td></tr></table></td>")
		fmt.Fprintf(w, "</tr>")
	}

	fmt.Fprintf(w, "</tbody></table>")

	//fmt.Fprintf(w, "<h3>Scan Config</h3>")
	//data, _ := ioutil.ReadFile("settings.scan")
	//fmt.Fprintf(w, "<code><pre>%v</pre></code>", string(data))

	fmt.Fprintf(w, webPageBottomTemplate)
}

func getScreenShots(w http.ResponseWriter, req *http.Request) {
	ip, _ := getIP(req)
	log.Println("getScreenShots Called by " + ip)
	fmt.Fprintf(w, "Scan is starting.....\n")
	//read database for HTTP/HTTPS/STUFF LIKE THAT
	//query open ports for host
	var scan []Scan
	db.Distinct("host", "protocol", "port").Order("host").Find(&scan)

	for _, host := range scan {

		//make hotlinks for known ports.
		value, ok := knownPortURLs[host.Port]

		// Make a Regex to say we only want letters and numbers
		reg, err := regexp.Compile("[^a-zA-Z0-9]+")
		if err != nil {
			log.Fatal(err)
		}
		processedString := reg.ReplaceAllString(value, "")

		//if http or https we do something
		url := host.Port
		if ok {
			if processedString == "http" || processedString == "https" {
				//call gowitness on URL
				log.Println("gowitness: " + value + host.Host + ":" + url)

				_, err := exec.Command("gowitness", "--delay=3", "single", (value + host.Host + ":" + url)).Output()
				if err != nil {
					log.Fatal(err.Error())
				}

			}
		}

	}

	fmt.Fprintf(w, "Get Screen Shots Completed.")
	//pass to gowitness

	//generate images

}

func getIP(r *http.Request) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("No valid ip found")
}
