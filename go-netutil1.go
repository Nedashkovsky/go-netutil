 

package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	"bufio"
	"os"
	"strconv"
	"strings"
)

var T = flag.Float64("t", 2, "update time(s)")
var C = flag.Uint("c", 0, "count (0 == unlimit)")
var Inter = flag.String("i", "*", "interface")
var U = flag.Float64("u", 99, "utilization")

var verbosity = flag.Int("v", 2, "verbosity")

type NetStat struct {
	Dev  []string
	Stat map[string]*DevStat
}

type DevStat struct {
	Name string
	Rx   uint64
	Tx   uint64
	BW   uint64
	Out  uint64
}


func FileExists(filename string) error {
        f, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
//        f, err := os.Open(filename)
        if err != nil {
                return "", err
        }
        defer f.Close()

}

func ReadLine1(filename string) (string, error) {
        f, err := os.Open(filename)
        if err != nil {
                return "", err
        }
        defer f.Close()

        var ret string

        r := bufio.NewReader(f)
        for {
                line, err := r.ReadString('\n')
                if err != nil {
                        break
                }
                ret = strings.Trim(line, "\n")
        }
        return ret, nil
}

func ReadLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}
	return ret, nil
}

// func getInfo() (ret NetStat) {

func (ret *NetStat) getInfo() error {
	lines, _ := ReadLines("/proc/net/dev")

	ret.Dev = make([]string, 0)
	ret.Stat = make(map[string]*DevStat)

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.Fields(strings.TrimSpace(fields[1]))

		// Vlogln(5, key, value)

		if *Inter != "*" && *Inter != key {
			continue
		}

		c := new(DevStat)
		// c := DevStat{}
		c.Name = key
		bw_name := "/sys/class/net/" + key + "/speed"
		wb1, _ := ReadLine1(bw_name)
		r1, err := strconv.ParseInt(wb1, 10, 64)
		if err != nil {
		    Vlogln(4, err)
		}
		c.BW = uint64(r1)

		r, err := strconv.ParseInt(value[0], 10, 64)
		if err != nil {
			Vlogln(4, key, "Rx", value[0], err)
			break
		}
		c.Rx = uint64(r)

		t, err := strconv.ParseInt(value[8], 10, 64)
		if err != nil {
			Vlogln(4, key, "Tx", value[8], err)
			break
		}
		c.Tx = uint64(t)
		
		ret.Dev = append(ret.Dev, key)
		ret.Stat[key] = c
		// Vlogln(1, ret.Stat[key])
	}

	return nil
}

func (ret *NetStat) getDiff( stat0 *NetStat, delta *NetStat) error { 
                for _, value := range ret.Dev {
                        t0, ok := stat0.Stat[value]
                        // fmt.Println("k:", key, " v:", value, ok)
                        if ok { 
                                dev, ok := delta.Stat[value]
                                if !ok {
                                        delta.Stat[value] = new(DevStat)
                                        dev = delta.Stat[value]
                                        delta.Dev = append(delta.Dev, value)
                                }
                                t1 := ret.Stat[value]
                                dev.Rx = t1.Rx - t0.Rx
                                dev.Tx = t1.Tx - t0.Tx
                                dev.BW = t1.BW
                                dev.Out = t1.Out
                        }
                }

                return nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime)
	flag.Parse()

	// runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.GOMAXPROCS(1)

	var stat0 NetStat
	var stat1 NetStat
//	stat0 := new(NetStat)
//	stat1 := new(NetStat)
	var delta NetStat
	delta.Dev = make([]string, 0)
	delta.Stat = make(map[string]*DevStat)
	
	name := "/opt/trafficserver/etc/trafficserver/healthchecks.bandwith"
	uid := 500
	gid := 500
	
	i := *C
	if i > 0 {
		i += 1
	}

	if *T < 0.01 {
		*T = 0.01
	}
	if *U > 99 {
		*U = 99
	}
	util := *U
	
	start := time.Now()
	elapsed := time.Since(start)
	if *Inter == "*" {
		fmt.Printf("\033c")
	}
	fmt.Printf("iface\t%-10s\tTx\n", "Rx")
	for {

		elapsed = time.Since(start)

		stat1.getInfo()
		stat1.getDiff( &stat0, &delta)
		stat0 = stat1

		// Vlogln(1, stat0)
		
		multi := len(delta.Dev)
//		if multi > 1 {
//			for i := 0; i < multi; i++ {
//				fmt.Printf("\033[%d;0H                                                        \r", i)
//			}
			fmt.Printf("\033[0;0Hiface\t%-10s\tTx\n", "Rx")
//			fmt.Printf("\033[2;0H")
//		}
		for _, iface := range delta.Dev {
			stat := delta.Stat[iface]
			if multi > 1 {
				if stat.BW != 0 {
				result1 := Vsize1(stat.Tx, *T, stat.BW, util)
				fmt.Printf("%v\t%v\t%v\t%v\t%v\n", iface, Vsize(stat.Rx, *T), Vsize(stat.Tx, *T), stat.BW, result1)
//				hc_file(result1)
				}
			} else {
				if stat.BW != 0 {
				result1 := Vsize1(stat.Tx, *T, stat.BW, util)
				fmt.Printf("\r%v\t%v\t%v\t%v\t%v", iface, Vsize(stat.Rx, *T), Vsize(stat.Tx, *T), stat.BW, result1)
				hc_file(result1, name, uid, gid)
				}
			}
		}
		// elapsed := time.Since(start)
		Vlogf(5, "[delta] %s", elapsed)
		start = time.Now()

		i -= 1
		if i == 0 {
			break
		}

		time.Sleep(time.Duration(*T*1000) * time.Millisecond)

	}
	if *Inter != "*" {
		fmt.Println()
	}
}

func Vsize(bytes uint64, delta float64) (ret string) {
	var tmp float64 = float64(bytes) / delta * 8
	var s string = " "

	bytes = uint64(tmp)

	switch {
	case bytes < uint64(2<<9):

	case bytes < uint64(2<<19):
		tmp = tmp / float64(2<<9)
		s = "K"

	case bytes < uint64(2<<29):
		tmp = tmp / float64(2<<19)
		s = "M"

	case bytes < uint64(2<<39):
		tmp = tmp / float64(2<<29)
		s = "G"

	case bytes < uint64(2<<49):
		tmp = tmp / float64(2<<39)
		s = "T"

	}
	ret = fmt.Sprintf("%06.2f %sB/s", tmp, s)
	return
}

func Vsize1(bytes uint64, delta float64, bw uint64, util float64) (ret string) {
        var tmp float64 = float64(bytes) / delta * 8

        percent := (tmp / (float64(bw) * 1000000)) * 100
        control := "OVERLOADED"
        if ( percent < util ){
    	    control = "OK"
        }
        
        ret = fmt.Sprintf("%s %6.2f", control, percent)
        return ret
}

func hc_file(hc, name string, uid, gid int) (err error) {
    fo, err := os.Create(name)
    if err != nil {
        panic(err)
    }
    // close fo on exit and check for its returned error
    defer func() {
        if err := fo.Close(); err != nil {
            panic(err)
        }
    }()
    // make a write buffer
    w := bufio.NewWriter(fo)
    _, err = w.WriteString(hc)
    err = os.Chown(name, uid, gid)
//    fmt.Printf("wrote %d bytes\n", n4)
    w.Flush()
    return nil
}

func Vlogf(level int, format string, v ...interface{}) {
	if level <= *verbosity {
		log.Printf(format, v...)
	}
}
func Vlog(level int, v ...interface{}) {
	if level <= *verbosity {
		log.Print(v...)
	}
}
func Vlogln(level int, v ...interface{}) {
	if level <= *verbosity {
		log.Println(v...)
	}
}



