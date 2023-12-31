package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb/v3"
	"github.com/dustin/go-humanize"
)

const (
	EMOJI_CROSSMARK = '\U0000274c'
)

// ///////////////////////////////////
// deal with command line arguments //
// ///////////////////////////////////
type CmdOptions struct {
	mode             string
	skipConfirmation bool
	fileURL          string
	outputFileName   string
}

func parseOptions() CmdOptions {
	var opts CmdOptions
	flag.StringVar(&opts.mode, "mode", "download", "[Optional] operation to perform. should be either \"download\" (download the file) or\n \"show\" (show information about the file).")
	flag.BoolVar(&opts.skipConfirmation, "y", false, "[Optional] when supplied, skip confirmation before starting the download.")
	flag.StringVar(&opts.fileURL, "url", "", "<Required> the URL you can retrieve on Google Drive by \"Share\" => \"Copy link\".")
	flag.StringVar(&opts.outputFileName, "o", "", "[Optional] filename to save the file as.")
	flag.Parse()
	return opts
}

// ///////////////////////////////
// deal with HTTP communication //
// ///////////////////////////////
type FileInfo struct {
	filesize uint64
	filename string
}

func extractInfoFromHeader(header http.Header) (FileInfo, error) {
	// retrieve file size
	fileSize, err := strconv.ParseUint(header.Get("Content-Length"), 10, 64)
	if err != nil {
		return FileInfo{}, fmt.Errorf("[ %c Error] Something went wrong during parsing Content-Length header: "+err.Error(), EMOJI_CROSSMARK)
	}
	// retrieve filename
	contentDisposeHeader := header.Get("Content-Disposition")
	re := regexp.MustCompile(`filename\*=UTF-8''(.+)`)
	fileNameKV := re.FindStringSubmatch(contentDisposeHeader)
	if len(fileNameKV) != 2 {
		return FileInfo{}, fmt.Errorf("[ %c Error] Something went wrong during parsing Content-Disposition header: "+err.Error(), EMOJI_CROSSMARK)
	}
	fileName, err := url.QueryUnescape(strings.TrimSpace(fileNameKV[1])) // filenameKV is an array like [filename="myfile.zip" myfile.zip], and these elements can be percent-encoded.
	if err != nil {
		return FileInfo{}, fmt.Errorf("[ %c Error] Something went wrong during decoding url-encoded filename: "+err.Error(), EMOJI_CROSSMARK)
	}

	return FileInfo{filesize: fileSize, filename: fileName}, nil
}

func showFileInfo(downloadURL string) error {
	resp, err := http.Head(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[ %c Error] Got an unusual response. Status code: %d", EMOJI_CROSSMARK, resp.StatusCode)
	}

	fileInfo, err := extractInfoFromHeader(resp.Header)
	if err != nil {
		return err
	}

	fmt.Printf(
		"[Information of the file]\n"+
			"filename: %s\n"+
			"filesize: %s\n",
		fileInfo.filename, humanize.IBytes(fileInfo.filesize),
	)
	return nil
}

func downloadFile(downloadURL, outputFileName string, skipConfirmation bool) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[ %c Error] Got an unusual response. Status code: %d", EMOJI_CROSSMARK, resp.StatusCode)
	}

	fileInfo, err := extractInfoFromHeader(resp.Header)
	if err != nil {
		return err
	}

	fileName := fileInfo.filename
	if outputFileName != "" { // if a filename is specified
		fileName = outputFileName
	}

	if !skipConfirmation {
		fmt.Printf("Download %s (%s) and save as %s? [Y/n]: ", fileInfo.filename, humanize.IBytes(fileInfo.filesize), fileName)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		if scanner.Text() == "n" {
			fmt.Println("Abort.")
			os.Exit(0)
		}
	}

	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// set up a progress bar and Reader associated to it
	reader := io.LimitReader(resp.Body, int64(fileInfo.filesize))
	progressbar := pb.Start64(int64(fileInfo.filesize))
	defer progressbar.Finish()
	barReader := progressbar.NewProxyReader(reader)

	if _, err := io.Copy(outFile, barReader); err != nil {
		return err
	}

	return nil
}

func main() {

	// parse command line argument and put them into a struct
	opts := parseOptions()

	// flag package doesn't allow us to mix positional and optional arguments.
	// in order not to confuse users, report an error if any positional argument is supplied.
	afterFlag := false
	for _, arg := range os.Args[1:] {
		if arg[0] == '-' {
			afterFlag = true
			continue
		}
		if !afterFlag { // found an positional argument that doesn't follow -flag
			fmt.Fprintf(os.Stderr, "[ %c Error] All arguments should be in \"-key value\" style.\n", EMOJI_CROSSMARK)
			flag.Usage()
			os.Exit(1)
		} else {
			afterFlag = false
		}
	}

	// check if a correct URL is supplied, and extract fileID to construct download URL
	re := regexp.MustCompile(`https://drive\.google\.com/file/d/([^/]{33})`)
	matches := re.FindStringSubmatch(opts.fileURL)
	if len(matches) != 2 {
		fmt.Fprintf(
			os.Stderr,
			"[ %c Error] The URL is not given or not in correct format. "+
				"Ensure that the url is like https://drive.google.com/file/d/<file ID as 33 characters>/view?usp=drive_link\n",
			EMOJI_CROSSMARK,
		)
		flag.Usage()
		os.Exit(1)
	}
	fileId := matches[1]
	// the value for "confirm" parameter seems to be arbitrary
	downloadURL := fmt.Sprintf("https://drive.google.com/uc?export=download&confirm=yes&id=%s", fileId)

	// perform operations based on mode parameter
	switch opts.mode {
	case "show":
		if err := showFileInfo(downloadURL); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	case "download":
		if err := downloadFile(downloadURL, opts.outputFileName, opts.skipConfirmation); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	default:
		fmt.Fprintf(os.Stderr, "[ %c Error] Unknown mode: %s\n", EMOJI_CROSSMARK, opts.mode)
		flag.Usage()
		os.Exit(1)
	}

}
