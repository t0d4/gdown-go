# gdown-go: Yet-another CLI downloader for Google Drive

![release workflow status](https://github.com/t0d4/gdown-go/actions/workflows/release.yml/badge.svg)

## About

This is a CLI tool to easily download large files from Google Drive.

## Installation

Download the latest binary for your platform from [the release page](https://github.com/t0d4/gdown-go/releases/latest) and give it executable permission.

## Usage

### Basic Usage

To download a file, run

```bash
$ gdown -url <sharing URL of the file to download>
```

where \<sharing URL of the file to download\> is the URL you can get on Google Drive by "Share" => "Copy link"

### Parameters

```bash
-mode string
    [Optional] operation to perform. (default: download)
    - "download" (download the file)
    - "show" (show information about the file)
-url string
    <Required> the URL you can retrieve on Google Drive by "Share" => "Copy link".
-o string
    [Optional] filename to save the file as.
-y
    [Optional] when supplied, skip confirmation before starting the download.
```

### Examples

- show information about the file

    ```bash
    $ gdown -mode show -url https://drive.google.com/file/d/1Joa7nl6y1FVBoBx1lSEJd-we86qcj7JA/view?usp=drive_link
    ```

- download the file as "myfile.zip" without confirmation

    ```bash
    $ gdown -y -url https://drive.google.com/file/d/1Joa7nl6y1FVBoBx1lSEJd-we86qcj7JA/view?usp=drive_link -o myfile.zip
    ```

## Note

If the file is larger than a certain size (exact size is not clear), Google Drive will show a prompt like "Can't scan for viruses" and ask us whether we want to download the file anyway. Currently, it seems that we can skip this prompt by including `confirm=<arbitrary string>` in the URL query string, but this behavior may be subject to change.
In case this tool stops working, feel free to contact me.