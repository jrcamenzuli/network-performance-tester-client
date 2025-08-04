package tests

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/jrcamenzuli/network-performance-tester-client/model"
	"github.com/jrcamenzuli/network-performance-tester-client/util"
)

const chunkSize = 10000000
const countBytesTransfer = 100000000 // 100MB

func DownloadThroughputTest(serverProtocol string, serverHost string, serverPort uint, pid uint) (model.ThroughputTest, error) {
	return DownloadThroughputTestWithProcesses(serverProtocol, serverHost, serverPort, pid, nil)
}

func DownloadThroughputTestWithProcesses(serverProtocol string, serverHost string, serverPort uint, pid uint, processNames []string) (model.ThroughputTest, error) {
	url := fmt.Sprintf("%s%s:%d/download/%d", serverProtocol, serverHost, serverPort, countBytesTransfer)

	client := util.CreateHTTPSClient()
	resp, err := client.Get(url)
	if err == nil {
		defer resp.Body.Close()
	} else {
		return model.ThroughputTest{Type: model.RX}, err
	}

	// Start monitoring processes if provided
	var processMonitoringDone sync.WaitGroup
	var processUsage model.ProcessCpuAndRam

	if len(processNames) > 0 {
		processMonitoringDone.Add(1)
		go func() {
			defer processMonitoringDone.Done()
			// Monitor for a reasonable throughput test duration (10 seconds should be enough)
			processUsage = util.MonitorProcessesContinuously(processNames, 10*time.Second, 100*time.Millisecond)
		}()
	}

	countBytesToReceive := countBytesTransfer
	var bytes []byte = make([]byte, chunkSize)
	tStart := time.Now()
	for countBytesToReceive > 0 {
		countBytesRead, err := resp.Body.Read(bytes)
		countBytesToReceive -= countBytesRead
		if err != nil {
			break
		}
	}
	tStop := time.Now()
	dt := tStop.Sub(tStart)
	countBytesTransferred := uint64(countBytesTransfer - countBytesToReceive)

	// Stop process monitoring
	if len(processNames) > 0 {
		processMonitoringDone.Wait()
	}

	cpuAndRam := util.GetCPUandRAM(pid)
	return model.ThroughputTest{
		Type:                  model.RX,
		CountBytesTransferred: countBytesTransferred,
		DurationNanoseconds:   uint64(dt.Nanoseconds()),
		CpuAndRam:             *cpuAndRam,
		ProcessCpuAndRam:      processUsage,
	}, nil
}

func UploadThroughputTest(serverProtocol string, serverHost string, serverPort uint, pid uint) (model.ThroughputTest, error) {
	return UploadThroughputTestWithProcesses(serverProtocol, serverHost, serverPort, pid, nil)
}

func UploadThroughputTestWithProcesses(serverProtocol string, serverHost string, serverPort uint, pid uint, processNames []string) (model.ThroughputTest, error) {
	url := fmt.Sprintf("%s%s:%d/upload", serverProtocol, serverHost, serverPort)
	countBytesToSend := countBytesTransfer
	var tStart time.Time
	var tStop time.Time
	countBytesSent := 0

	//buffer for storing multipart data
	byteBuf := &bytes.Buffer{}

	mpWriter := multipart.NewWriter(byteBuf)

	//part: file
	mpWriter.CreateFormFile("file", "")
	contentType := mpWriter.FormDataContentType()

	nmulti := byteBuf.Len()
	multi := make([]byte, nmulti)
	_, _ = byteBuf.Read(multi)

	//part: latest boundary
	//when multipart closed, latest boundary is added
	mpWriter.Close()
	nboundary := byteBuf.Len()
	lastBoundary := make([]byte, nboundary)
	_, _ = byteBuf.Read(lastBoundary)

	//calculate content length
	totalSize := int64(nmulti) + countBytesTransfer + int64(nboundary)

	// Start monitoring processes if provided
	var processMonitoringDone sync.WaitGroup
	var processUsage model.ProcessCpuAndRam

	if len(processNames) > 0 {
		processMonitoringDone.Add(1)
		go func() {
			defer processMonitoringDone.Done()
			// Monitor for a reasonable throughput test duration (10 seconds should be enough)
			processUsage = util.MonitorProcessesContinuously(processNames, 10*time.Second, 100*time.Millisecond)
		}()
	}

	//use pipe to pass request
	rd, wr := io.Pipe()
	defer rd.Close()

	go func() {
		defer wr.Close()

		//write multipart
		_, _ = wr.Write(multi)

		buff := make([]byte, chunkSize)

		for countBytesSent < countBytesTransfer {
			if countBytesToSend < chunkSize {
				n, _ := wr.Write(buff[:countBytesToSend])
				countBytesToSend -= n
				countBytesSent += n
				break
			} else {
				n, err := wr.Write(buff)
				countBytesToSend -= n
				countBytesSent += n
				if err != nil {
					break
				}
			}
		}
		//write boundary
		_, _ = wr.Write(lastBoundary)
	}()

	//construct request with rd
	req, _ := http.NewRequest("POST", url, rd)
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = totalSize

	//process request
	client := util.CreateHTTPSClient()
	tStart = time.Now()
	resp, err := client.Do(req)
	tStop = time.Now()
	if err != nil {
		log.Fatal(err)
	} else {
		resp.Body.Close()
	}

	dt := tStop.Sub(tStart)

	// Stop process monitoring
	if len(processNames) > 0 {
		processMonitoringDone.Wait()
	}

	cpuAndRam := util.GetCPUandRAM(pid)
	return model.ThroughputTest{
		Type:                  model.TX,
		CountBytesTransferred: uint64(countBytesSent),
		DurationNanoseconds:   uint64(dt.Nanoseconds()),
		CpuAndRam:             *cpuAndRam,
		ProcessCpuAndRam:      processUsage,
	}, nil
}
