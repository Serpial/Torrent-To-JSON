package main

import (
	"fmt"
	"os"
	"encoding/json"
	"time"
	"io/ioutil"
	"strings"
	"net/http"

	"github.com/j-muller/go-torrent-parser"
)

type ArgumentsFormat struct {
	InputLoc string
	OutputLoc string
	WillOutput bool
	ValidInput bool
	IsHelp bool
	DontServe bool
}

type Torrent struct {
	CreatedBy string
	CreatedAt time.Time
	InfoHash string
	Comment string
	Files []*gotorrentparser.File
	Announce []string
}

func main() {
	args := verArgs()

	if (args.ValidInput && !args.IsHelp) {
		// Get and parse Torrent file
		parsedTorrent, err := gotorrentparser.ParseFromFile(args.InputLoc)
		if err == nil {
			torrent, _ := json.MarshalIndent(
				&Torrent{
					CreatedBy: parsedTorrent.CreatedBy,
					CreatedAt: parsedTorrent.CreatedAt,
					InfoHash: parsedTorrent.InfoHash,
					Comment: parsedTorrent.Comment,
					Files: parsedTorrent.Files,
					Announce: parsedTorrent.Announce,
				}, "", "\t",
			)

			if (args.WillOutput) {
				if (writeOut(args, string(torrent)) == nil) {
					fmt.Println("Success: Written to file")
				} else {
					fmt.Println("Error: Could not write out file")
				}
			}
			if (args.DontServe) {
				fmt.Println(string(torrent))
			} else {
				fmt.Println("Your file is being served to: http://localhost:8080/")
				http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
					fmt.Fprintf(w, string(torrent))
				})
				http.ListenAndServe(":8080", nil)
			}
		} else {
			fmt.Println(err)
		}
	} else if (!args.ValidInput) {
		fmt.Println("Error: Poorly formatted command")
	} else if (args.IsHelp) {
		fmt.Println(helpString())
	}
}

func writeOut(args *ArgumentsFormat, tor string) error {
	var err error
	
	file ,_ := os.Stat(args.OutputLoc)

	if file == nil || file.Mode().IsRegular() {
		err = ioutil.WriteFile(
			args.OutputLoc,
			[]byte(tor),
			0644,
		)
	} else if (file.Mode().IsDir()) {
		originalFile := args.InputLoc[strings.LastIndex(args.InputLoc, "/")+1:]
		nameMinusExtension := originalFile[:strings.LastIndex(originalFile, ".")]
		
		err = ioutil.WriteFile(
			args.OutputLoc+(nameMinusExtension+".json"),
			[]byte(tor),
			0644,
		)
	} 
	return err
}

func helpString() string {
	outString :=
		"Program to decode a Bencoded file. Make sure that\n"+
			"the input file is last in the command\n" +
			"\t'-o'\t\t\tOutputs the file using the same name as the parent to the local directory\n"+
			"\t'-o <Location>'\tOutputs the file in JSON to a specific location\n"+
			"\t'-N'\t\t\tOutputs JSON formatted code to console instead of hosting a webpage\n"+
			"\t'--help'\t\tGet this help screen"
	return outString 
}

func verArgs() *ArgumentsFormat {
	args := os.Args[1:]
	identifiers := map[string]bool {
		"o" : true, // Output to json to location
		"N" : true, // Output to console instead of serving locally
		"-help" : true,
	}
	structuredArgs := ArgumentsFormat{
		OutputLoc:"./",
		WillOutput:false,
		ValidInput:true,
		IsHelp:false,
		DontServe:false,
	}

	if (strings.Contains(strings.Join(args, " "), "--help")) {
		structuredArgs.IsHelp = true;
		return &structuredArgs
	}
	
	if (len(args) > 0 && !identifiers[args[len(args)-1][1:]] ) {
		structuredArgs.InputLoc = args[len(args)-1]
		for i := 0; i<len(args)-1; i++ {
			if (identifiers[args[i][1:]]) {
				if (args[i] == "-o" && i < len(args)-1) {
					
					structuredArgs.WillOutput = true
					if (i < len(args)-2 && !identifiers[args[i+1][1:]]){
						structuredArgs.OutputLoc = args[i+1]
						i += 1
					}					
				} else if (args[i] == "-N"){
					
					structuredArgs.DontServe = true
				} else { structuredArgs.ValidInput = false }
			} else { structuredArgs.ValidInput = false }
		}
	} else { structuredArgs.ValidInput = false }
	
	return &structuredArgs
}
