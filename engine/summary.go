// Copyright 2024 Christoph Fichtmüller. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package engine

import (
	"slices"
	"sort"
	"strings"
	"time"
)

type SummaryStats struct {
	Endpoint                   string
	Availability               float64
	AvgResponseTime            time.Duration
	MedianResponseTime         time.Duration
	Percentile99ResponseTime   time.Duration
	LongestResponseTime        time.Duration
	ShortestCertValidityTime   time.Duration
	WorstMonitor               string
	NumberOfMeasurements       int
	NumberOfFailedMeasurements int
	MonitoringDuration         string
}

func Summarize(pings []*Ping) []*SummaryStats {
	index := make(map[string]*SummaryStats)
	endpointsData := make(map[string][]*Ping)

	// Group pings by endpoint
	for _, p := range pings {
		endpointsData[p.URL] = append(endpointsData[p.URL], p)
	}

	// Calculate statistics per endpoint
	for endpoint, data := range endpointsData {
		var totalResponseTime, successCount, longestResponseTime, shortestCertValidity, failedCount int
		var responseTimes []int
		shortestCertValidity = int(^uint(0) >> 1) // Set to max int initially
		var worstMonitorName string
		worstPerformance := 0

		for _, p := range data {
			pTotalResponseTime := int(p.TotalResponseTime.Milliseconds())
			totalResponseTime += pTotalResponseTime
			responseTimes = append(responseTimes, pTotalResponseTime)
			if p.Status == "Success" {
				successCount++
			} else {
				failedCount++
			}
			if int(p.TotalResponseTime) > longestResponseTime {
				longestResponseTime = pTotalResponseTime
			}
			if int(p.CertRemainingValidity) < shortestCertValidity {
				shortestCertValidity = int(p.CertRemainingValidity.Seconds())
			}
			// Determine the worst monitor based on response time
			if int(p.TotalResponseTime) > worstPerformance {
				worstPerformance = pTotalResponseTime
				worstMonitorName = p.Name
			}
		}

		// Sort response times to calculate median and 99th percentile
		sort.Ints(responseTimes)
		var medianResponseTime float64
		var percentile99ResponseTime int
		if len(responseTimes) > 0 {
			medianResponseTime = float64(responseTimes[len(responseTimes)/2])
			percentileIndex := int(float64(len(responseTimes))*0.99) - 1
			if percentileIndex < 0 {
				percentileIndex = 0
			}
			percentile99ResponseTime = responseTimes[percentileIndex]
		}

		// Calculate availability
		availability := (float64(successCount) / float64(len(data))) * 100

		// Calculate average response time
		avgResponseTime := float64(totalResponseTime) / float64(len(data))

		// Determine monitoring duration
		monitoringDuration := "unknown" // Placeholder; calculation can be done based on timestamps if available

		// Store stats
		index[endpoint] = &SummaryStats{
			Endpoint:                   endpoint,
			Availability:               availability,
			AvgResponseTime:            time.Duration(avgResponseTime) * time.Millisecond,
			MedianResponseTime:         time.Duration(medianResponseTime) * time.Millisecond,
			Percentile99ResponseTime:   time.Duration(percentile99ResponseTime) * time.Millisecond,
			LongestResponseTime:        time.Duration(longestResponseTime) * time.Millisecond,
			ShortestCertValidityTime:   time.Duration(shortestCertValidity) * time.Millisecond,
			WorstMonitor:               worstMonitorName,
			NumberOfMeasurements:       len(data),
			NumberOfFailedMeasurements: failedCount,
			MonitoringDuration:         monitoringDuration,
		}
	}

	stats := make([]*SummaryStats, 0, len(index))
	for _, s := range index {
		stats = append(stats, s)
	}

	slices.SortFunc(stats, func(a, b *SummaryStats) int {
		return strings.Compare(a.Endpoint, b.Endpoint)
	})

	return stats
}
