// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Monitor defines what and how to monitor
type Monitor struct {
	Name                string
	URL                 string
	Retries             int
	RetryInterval       int
	ConnectTimeout      time.Duration
	ResponseTimeout     time.Duration
	MaxRedirects        int
	AcceptedStatusCodes []int
	HTTPMethod          string
	Headers             map[string]string
}

// Ping is the result of a monitoring event
type Ping struct {
	Name                  string
	URL                   string
	Status                string
	Timestamp             time.Time
	StatusCode            int
	Message               string
	DNSTime               time.Duration
	ConnectionTime        time.Duration
	TLSTime               time.Duration
	TTFB                  time.Duration
	DownloadTime          time.Duration
	TotalResponseTime     time.Duration
	CertRemainingValidity time.Duration
}

func main() {
	file := flag.String("f", "", "Path to the file")
	fname := flag.String("n", "", "Name of the monitor")
	flag.Parse()
	if *file == "" && len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "Usage: monitor [URL]...\n")
		os.Exit(1)
	}

	name := *fname
	if name == "" {
		n, err := os.Hostname()
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to determine hostname: %v\n", err)
			os.Exit(1)
		}
		name = n
	}

	writer := NewCsvWriter(os.Stdout)
	wait := &sync.WaitGroup{}
	urls := os.Args[1:]

	if *file != "" {
		if len(os.Args) < 3 {
			fmt.Fprint(os.Stderr, "Usage: monitor -f [file]\n")
			os.Exit(1)
		}
		b, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read file %s: %v\n", os.Args[2], err)
			os.Exit(1)
		}
		urls = strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	}

	invalid := false
	for _, ru := range urls {
		if ru == "" {
			continue
		}
		u, err := url.Parse(ru)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid url '%s': %v\n", ru, err)
			invalid = true
		} else if u.Scheme != "http" && u.Scheme != "https" {
			fmt.Fprintf(os.Stderr, "Invalid url '%s'\n", ru)
			invalid = true
		}
	}
	if invalid {
		os.Exit(1)
	}

	for _, u := range urls {
		if u == "" {
			continue
		}
		wait.Add(1)
		go pingUrl(writer, wait, name, u)
	}

	wait.Wait()
	writer.Flush()
}

func pingUrl(w Writer, wg *sync.WaitGroup, name, url string) {
	monitor := &Monitor{
		Name:                name,
		URL:                 url,
		Retries:             3,
		RetryInterval:       10,
		ConnectTimeout:      5 * time.Second,
		ResponseTimeout:     10 * time.Second,
		MaxRedirects:        3,
		AcceptedStatusCodes: []int{200, 201, 202, 204},
		HTTPMethod:          "GET",
		Headers:             map[string]string{"User-Agent": "HTTP-Monitor-Agent"},
	}
	ping := executePing(monitor)

	w.Write(ping)
	wg.Done()
}

// executePing takes a Monitor and produces a Ping
func executePing(monitor *Monitor) *Ping {
	// Timing variables
	var dnsStart, connStart, tlsStart, firstByteTime time.Time
	var dnsDuration, connDuration, tlsDuration, downloadTime time.Duration
	var certRemainingValidity time.Duration

	// Create a custom HTTP transport with separate connect and response timeouts
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: monitor.ConnectTimeout,
		}).DialContext,
		TLSHandshakeTimeout: monitor.ConnectTimeout, // Apply the connect timeout to the TLS handshake
	}

	// Create a custom HTTP client
	client := &http.Client{
		Transport: transport,
		Timeout:   monitor.ResponseTimeout,
	}

	// Create an HTTP request with the appropriate method and headers
	req, err := http.NewRequest(monitor.HTTPMethod, monitor.URL, nil)
	if err != nil {
		return &Ping{
			Name:      monitor.Name,
			URL:       monitor.URL,
			Status:    "Failed",
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Error creating request: %v", err),
		}
	}

	for key, value := range monitor.Headers {
		req.Header.Set(key, value)
	}

	// Add trace to measure DNS, connection, TLS handshake times, and TTFB
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsDuration = time.Since(dnsStart)
		},
		ConnectStart: func(network, addr string) {
			connStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connDuration = time.Since(connStart)
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			tlsDuration = time.Since(tlsStart)
			if err == nil {
				// If TLS handshake succeeded, check the certificate validity
				if len(state.PeerCertificates) > 0 {
					cert := state.PeerCertificates[0]
					remaining := time.Until(cert.NotAfter)
					certRemainingValidity = remaining
				}
			}
		},
		GotFirstResponseByte: func() {
			firstByteTime = time.Now()
		},
	}

	// Associate the trace with the request's context
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	// Record the start time of the request
	start := time.Now()

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return &Ping{
			Name:      monitor.Name,
			URL:       monitor.URL,
			Status:    "Failed",
			Timestamp: time.Now(),
			Message:   fmt.Sprintf("Error executing request: %v", err),
		}
	}
	defer resp.Body.Close()

	// Calculate TTFB
	ttfb := firstByteTime.Sub(start)

	// Measure download time (after the first byte)
	downloadStart := time.Now()
	_, _ = http.MaxBytesReader(nil, resp.Body, 10<<20).Read(make([]byte, 10<<20)) // Limiting to 10MB read for example
	downloadTime = time.Since(downloadStart)

	// Calculate total response time
	totalDuration := time.Since(start)

	// Determine if status code is accepted
	status := "Success"
	if !isStatusCodeAccepted(resp.StatusCode, monitor.AcceptedStatusCodes) {
		status = "Failed"
	}

	// Return the Ping result, including certRemainingValidity if it's a TLS connection
	return &Ping{
		Name:                  monitor.Name,
		URL:                   monitor.URL,
		Status:                status,
		Timestamp:             time.Now(),
		StatusCode:            resp.StatusCode,
		Message:               http.StatusText(resp.StatusCode),
		DNSTime:               dnsDuration,
		ConnectionTime:        connDuration,
		TLSTime:               tlsDuration,
		TTFB:                  ttfb,
		DownloadTime:          downloadTime,
		TotalResponseTime:     totalDuration,
		CertRemainingValidity: certRemainingValidity,
	}
}

func isStatusCodeAccepted(statusCode int, acceptedStatusCodes []int) bool {
	for _, code := range acceptedStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

type Writer interface {
	Write(ping *Ping) error
	Flush()
}

type CsvWriter struct {
	writer *csv.Writer
}

func NewCsvWriter(w io.Writer) *CsvWriter {
	writer := csv.NewWriter(w)
	writer.Comma = ';'

	return &CsvWriter{
		writer: writer,
	}
}

func (w *CsvWriter) Write(ping *Ping) error {
	return w.writer.Write([]string{
		ping.Name,
		ping.URL,
		ping.Status,
		ping.Timestamp.Format(time.RFC3339),
		strconv.Itoa(ping.StatusCode),
		ping.Message,
		strconv.FormatInt(ping.DNSTime.Milliseconds(), 10),
		strconv.FormatInt(ping.ConnectionTime.Milliseconds(), 10),
		strconv.FormatInt(ping.TLSTime.Milliseconds(), 10),
		strconv.FormatInt(ping.TTFB.Milliseconds(), 10),
		strconv.FormatInt(ping.DownloadTime.Milliseconds(), 10),
		strconv.FormatInt(ping.TotalResponseTime.Milliseconds(), 10),
		strconv.FormatInt(ping.CertRemainingValidity.Microseconds()/1000, 10),
	})
}

func (w *CsvWriter) Flush() {
	w.writer.Flush()
}
