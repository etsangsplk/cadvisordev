// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Usage
// This driver expects the following ENV variables:
// WAVEFRONT_INTERVAL=10 - the number of secodns between flushes to WF
// HOST_TAGS="az=\"us-west-2\" app=\"cadvisortesting\"" - Tags that will be added to all metrics
// ./cadvisor -storage_driver=wavefront -storage_driver_host=wavefront:2878 -storage_driver_db=$(hostname) --allow_dynamic_housekeeping=false --global_housekeeping_interval=30s --housekeeping_interval=10s

package wavefront

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	//"github.com/golang/glog"
	info "github.com/google/cadvisor/info/v1"
	"github.com/google/cadvisor/storage"
)

func init() {
	storage.RegisterStorageDriver("wavefront", new)
}

type wavefrontStorage struct {
	Source       string
	ProxyAddress string
	LastFlush    map[string]time.Time
	Conn         net.Conn
	WfInterval   int
	WfAddTags    string
	WfPrefix     string
}

const (
	colCpuCumulativeUsage = "cpu_cumulative_usage"
	// Memory Usage
	colMemoryUsage = "memory_usage"
	// Working set size
	colMemoryWorkingSet = "memory_working_set"
	// Cumulative count of bytes received.
	colRxBytes = "rx_bytes"
	// Cumulative count of receive errors encountered.
	colRxErrors = "rx_errors"
	// Cumulative count of bytes transmitted.
	colTxBytes = "tx_bytes"
	// Cumulative count of transmit errors encountered.
	colTxErrors = "tx_errors"
	// Filesystem summary
	colFsSummary = "fs_summary"
	// Filesystem limit.
	colFsLimit = "fs_limit"
	// Filesystem usage.
	colFsUsage = "fs_usage"
)

func new() (storage.StorageDriver, error) {

	return newStorage(
		*storage.ArgDbName,
		*storage.ArgDbHost,
	)
}

func (driver *wavefrontStorage) containerStatsToValues(stats *info.ContainerStats) (series map[string]uint64) {
	series = make(map[string]uint64)

	// Cumulative Cpu Usage
	series[colCpuCumulativeUsage] = stats.Cpu.Usage.Total

	// Memory Usage
	series[colMemoryUsage] = stats.Memory.Usage

	// Working set size
	series[colMemoryWorkingSet] = stats.Memory.WorkingSet

	// Network stats.
	series[colRxBytes] = stats.Network.RxBytes
	series[colRxErrors] = stats.Network.RxErrors
	series[colTxBytes] = stats.Network.TxBytes
	series[colTxErrors] = stats.Network.TxErrors

	return series
}

func (driver *wavefrontStorage) containerFsStatsToValues(series *map[string]uint64, stats *info.ContainerStats) {
	for _, fsStat := range stats.Filesystem {
		// Summary stats.
		(*series)[colFsSummary+"."+colFsLimit] += fsStat.Limit
		(*series)[colFsSummary+"."+colFsUsage] += fsStat.Usage

		// Per device stats.
		(*series)[fsStat.Device+"~"+colFsLimit] = fsStat.Limit
		(*series)[fsStat.Device+"~"+colFsUsage] = fsStat.Usage
	}
}

func (driver *wavefrontStorage) AddStats(ref info.ContainerReference, stats *info.ContainerStats) error {
	if stats == nil {
		return nil
	}

	//Container name
	containerName := ref.Name
	if len(ref.Aliases) > 0 {
		containerName = ref.Aliases[0]
	}
	//Only send to WF if the interval has passed for this container.
	current := time.Now()
	dur := current.Sub(driver.LastFlush[containerName])
	//Get the Wavefront interval variable
	//osInterval, err := strconv.Atoi(os.Getenv("WF_INTERVAL"))
	osInterval := driver.WfInterval
	interval := osInterval
	if dur.Seconds() < float64(interval) {
		//it's not time to flush, do nothing
		return nil
	}
	//glog.Info("Flushing container stats for " + containerName)
	driver.LastFlush[containerName] = time.Now()

	//Get current timestamp
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	//Source tag (from flag)
	source := driver.Source

	//See if additional host tags were passed
	addTags := driver.WfAddTags

	//Image


	//Additional tags (namespace and labels)
	appendTags := ""
	//Namespace
	ns := ref.Namespace
	if ns != "" {
		appendTags += " namespace=\"" + ns + "\""
	}
	//Labels
	labels := ref.Labels
	for key, value := range labels {
		appendTags += " " + key + "=\"" + value + "\""
	}

	//metric data
	series := driver.containerStatsToValues(stats)
	//metric data on volumes
	driver.containerFsStatsToValues(&series, stats)
	for key, value := range series {
		if strings.Contains(key, "~") {
			// storage device metrics - extract device as point tag.
			parts := strings.Split(key, "~")
			newKey := parts[1]
			//pointTagVal := strings.Replace(parts[0], "/", "-", -1)
			pointTagVal := parts[0]
			fmt.Fprintf(driver.Conn, fmt.Sprintf("%s%s %v %s source=%s container=\"%s\" device=\"%s\" %s %s \n", driver.WfPrefix, newKey, value, timestamp, source, containerName, pointTagVal, addTags, appendTags))
		} else {
			fmt.Fprintf(driver.Conn, fmt.Sprintf("%s%s %v %s source=%s container=\"%s\" %s %s \n", driver.WfPrefix, key, value, timestamp, source, containerName, addTags, appendTags))
		}
	}
	return nil
}

func (driver *wavefrontStorage) Close() error {
	driver.Conn.Close()
	return nil
}

func newStorage(source string, proxyAddress string) (*wavefrontStorage, error) {

	wavefrontStorage := &wavefrontStorage{
		Source:       source,
		ProxyAddress: proxyAddress,
		WfInterval:   0,
		WfAddTags:    "",
		WfPrefix:     "cadvisor.",
	}

	// Timeout if unable to connect after 10 seconds.
	conn, err := net.DialTimeout("tcp", proxyAddress, time.Second*10)
	wavefrontStorage.Conn = conn
	// Initialize map that will hold timestamp of the last flush for each container
	wavefrontStorage.LastFlush = make(map[string]time.Time)
	// Load environment variables
	if os.Getenv("WF_INTERVAL") != "" {
		wavefrontStorage.WfInterval, _ = strconv.Atoi(os.Getenv("WF_INTERVAL"))
	}
	if os.Getenv("WF_ADD_TAGS") != "" {
		wavefrontStorage.WfAddTags = os.Getenv("WF_ADD_TAGS")
	}
	if os.Getenv("WF_PREFIX") != "" {
		wavefrontStorage.WfPrefix = os.Getenv("WF_PREFIX")
	}
	return wavefrontStorage, err

}
