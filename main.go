package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type serverStat struct {
	CurrentLA             int //Текущий Load Average сервера.
	MemBytesAvailable     int //Текущий объём оперативной памяти сервера в байтах.
	MemBytesUsed          int //Текущее потребление оперативной памяти сервера в байтах.
	DiskBytesAvailable    int //Текущий объём дискового пространства сервера в байтах.
	DiskBytesUsed         int //Текущее потребление дискового пространства сервера в байтах.
	NetBandwidthAvailable int //Текущая пропускная способность сети в байтах в секунду.
	NetBandwidthUsed      int //Текущая загруженность сети в байтах в секунду.
}

const (
	unitB  = 1
	unitKb = unitB * 1024
	unitMb = unitKb * 1024
	unitGb = unitMb * 1024

	unitBps  = 1.0
	unitKbps = unitBps * 1000
	unitMbps = unitKbps * 1000
	unitGbps = unitMbps * 1000
)

func parseServerStat(input []byte) (serverStat, error) {
	parts := strings.Split(string(input), ",")

	if len(parts) != 7 {
		return serverStat{}, fmt.Errorf("expected 7 values, got %d", len(parts))
	}

	var stat serverStat

	var err error
	stat.CurrentLA, err = strconv.Atoi(parts[0])
	if err != nil {
		return stat, fmt.Errorf("error parsing CurrentLA: %v", err)
	}

	stat.MemBytesAvailable, err = strconv.Atoi(parts[1])
	if err != nil {
		return stat, fmt.Errorf("error parsing MemBytesAvailable: %v", err)
	}

	stat.MemBytesUsed, err = strconv.Atoi(parts[2])
	if err != nil {
		return stat, fmt.Errorf("error parsing MemBytesUsed: %v", err)
	}

	stat.DiskBytesAvailable, err = strconv.Atoi(parts[3])
	if err != nil {
		return stat, fmt.Errorf("error parsing DiskBytesAvailable: %v", err)
	}

	stat.DiskBytesUsed, err = strconv.Atoi(parts[4])
	if err != nil {
		return stat, fmt.Errorf("error parsing DiskBytesUsed: %v", err)
	}

	stat.NetBandwidthAvailable, err = strconv.Atoi(parts[5])
	if err != nil {
		return stat, fmt.Errorf("error parsing NetBandwidthAvailable: %v", err)
	}

	stat.NetBandwidthUsed, err = strconv.Atoi(parts[6])
	if err != nil {
		return stat, fmt.Errorf("error parsing NetBandwidthAvailable: %v", err)
	}

	return stat, nil
}

func checkServerStat(stat serverStat) {
	//При превышении значения 30 необходимо вывести в консоль сообщение
	//Load Average is too high: N, где N — текущее значение.
	if stat.CurrentLA > 30 {
		fmt.Printf("Load Average is too high: %d\n", stat.CurrentLA)
	}

	//При превышении 80% от имеющегося объёма необходимо вывести в консоль сообщение
	//Memory usage too high: N%, где N — текущее процентное значение.
	memUsagePercent := int(math.Round(float64(stat.MemBytesUsed) / float64(stat.MemBytesAvailable) * 100))
	if memUsagePercent > 80 {
		fmt.Printf("Memory usage is too high: %d%% \n", memUsagePercent)
	}

	//При превышении 90% от имеющегося объёма необходимо вывести в консоль сообщение
	//Free disk space is too low: N Mb left, где N — количество оставшихся свободных мегабайтов.
	diskUsagePercent := int(math.Round(float64(stat.DiskBytesUsed) / float64(stat.DiskBytesAvailable) * 100))
	if diskUsagePercent > 90 {
		fmt.Printf("Free disk space is too low: %d Mb left\n", (stat.DiskBytesAvailable-stat.DiskBytesUsed)/unitMb)
	}

	//При превышении 90% от имеющейся пропускной способности вывести в консоль сообщение
	//Network bandwidth usage high: N Mbit/s available, где N — размер свободной полосы в мегабитах в секунду.
	netUsagePercent := int(math.Round(float64(stat.NetBandwidthUsed) / float64(stat.NetBandwidthAvailable) * 100))
	if netUsagePercent > 90 {
		fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", (stat.NetBandwidthAvailable-stat.NetBandwidthUsed)/unitMbps)
	}
}

func main() {
	const maxRetries = 3
	var err error
	var response *http.Response

	for attempts := 0; attempts < maxRetries; attempts++ {
		response, err = http.Get("http://srv.msk01.gigacorp.local/_stats")
		if err != nil {
			fmt.Println("Error:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if response.StatusCode != http.StatusOK {
			fmt.Printf("Received non-200 response: %d\n", response.StatusCode)
			response.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			fmt.Println("Error reading response body:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		stat, err := parseServerStat(body)
		if err != nil {
			fmt.Println("Error parsing server statistics:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("Parsed serverStat: %+v\n", stat)
		checkServerStat(stat)
		return
	}

	fmt.Println("Unable to fetch server statistic")
}
