# 🌐 ssl_explorer

## Description
🔍 ssl_explorer is a tool designed for cybersecurity professionals and ethical hackers. It streamlines the extraction of SSL/TLS certificate information from specified URLs, crucial for target reconnaissance in security assessments and penetration testing.

## Features
- 🔐 Extracts SSL/TLS certificate information from URLs.
- 📁 Supports processing multiple URLs via an input file.
- 🌟 Offers single URL processing.
- ⚙️ Concurrent processing with customizable thread count.
- 📊 Outputs results in a readable CSV format.

## Installation
Install ssl_explorer using `go get`:
```
go get github.com/Chocapikk/ssl_explorer
```

## Usage
To use ssl_explorer, specify either a single URL or provide a file containing multiple URLs.

Single URL:
```
ssl_explorer -url=https://example.com
```

Multiple URLs from a file:
```
ssl_explorer -input=urls.txt
```

Specify the number of concurrent threads (default is 5):
```
ssl_explorer -input=urls.txt -threads=10
```

To save output to a file:
```
ssl_explorer -input=urls.txt -output=results.csv
```
