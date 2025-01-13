package main

import (
	"flag"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

const BYTE_IEC = 1024
const SHOW_NUM_PROCESS = 25
const SHOW_COUNT_PROCESS = 5
const LIMIT_COMMAND_CHAR = 150
const GROUP_SYSTEM_APP = true

func main() {
	showNumProcess := flag.Int("n", SHOW_NUM_PROCESS, "number process to show")
	showCountProcess := flag.Int("p", SHOW_COUNT_PROCESS, "number process to count each command")
	groupSystemApp := flag.Bool("s", GROUP_SYSTEM_APP, "group system app command")
	flag.Parse()

	output := RunCommand(`ps x -o %cpu,rss,command -m -A`)
	d, totalMemory, totalProcess := DisplayOutput(output, *showNumProcess, *showCountProcess, *groupSystemApp)
	fmt.Println("Total memory using", ByteCountIEC(totalMemory))
	fmt.Println("Total process", totalProcess)
	fmt.Println(d)
}

func RunCommand(command string) string {
	cmd := exec.Command(`bash`, `-c`, command)
	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return string(output)
}

type Process struct {
	CpuPercent float64
	Rss        int
	Comm       string
	TotalP     int
}

func DisplayOutput(output string, showNumProcess, showCountProcess int, groupSystemApp bool) (linesCombine string, totalMemory int, totalProcess int) {
	outputs := strings.Split(output, "\n")
	comms := []string{}

	firstLine := outputs[0]

	splitComm := strings.Split(firstLine, "COMM")
	beginComm := len(splitComm[0])

	splitCpu := strings.Split(firstLine, "%CPU")
	beginCpu := len(splitCpu[0]) + 4

	mapProccess := map[string]*Process{}
	lines := []string{}
	for i, line := range outputs {
		newline := line
		if i == 0 {
			newline = line[:beginCpu] + "  " + line[beginCpu:]
			lines = append(lines, newline)
			continue
		}
		if line == "" {
			continue
		}

		var cpuPercent float64
		var rssInt int
		var err error
		if len(line) > beginComm {
			{
				cpu := strings.TrimSpace(line[:beginCpu])
				cpuPercent, err = strconv.ParseFloat(cpu, 64)
				if err != nil {
					panic(err)
				}
			}
			{
				rss := strings.TrimSpace(line[beginCpu:beginComm])
				rssInt, err = strconv.Atoi(rss)
				if err != nil {
					panic(err)
				}
			}

			comm := strings.TrimSpace(line[beginComm:])
			trimComm := comm
			if strings.HasPrefix(comm, "/") {
				if strings.HasPrefix(comm, "/Applications/") {
					newComm := strings.TrimPrefix(comm, "/Applications/")
					newComms := strings.Split(newComm, "/")
					trimComm = newComms[0]
					trimComm = strings.TrimSuffix(trimComm, ".app")
					if trimComm == "" {
						trimComm = newComms[len(newComms)-1]
					}
				} else if groupSystemApp && (strings.HasPrefix(comm, "/System/Library/") ||
					strings.HasPrefix(comm, "/usr/libexec/") ||
					strings.HasPrefix(comm, "/usr/sbin/") ||
					strings.HasPrefix(comm, "/Library/")) {
					trimComm = "System"
				} else if strings.HasPrefix(comm, `/System/Applications/`) {
					newComm := strings.TrimPrefix(comm, "/System/Applications/")
					newComms := strings.Split(newComm, "/")
					trimComm = newComms[0]
					if len(newComms) > 1 && strings.HasSuffix(newComms[1], ".app") {
						trimComm = newComms[1]
					}

					if trimComm == "" {
						trimComm = newComms[len(newComms)-1]
					}
				} else if strings.HasPrefix(comm, "/System/") {
					newComms := strings.Split(comm, "/")
					trimComm = newComms[len(newComms)-1]
				}
			}
			newline = line[:beginComm] + trimComm

			if _, found := mapProccess[trimComm]; !found {
				mapProccess[trimComm] = &Process{
					CpuPercent: cpuPercent,
					Rss:        rssInt,
					Comm:       trimComm,
					TotalP:     1,
				}
				comms = append(comms, trimComm)
			} else {
				r := mapProccess[trimComm]
				r.CpuPercent += cpuPercent
				r.Rss += rssInt
				r.TotalP++
			}
		} else {
			fmt.Println("line", line)
		}
	}

	processList := SortProcess(mapProccess)

	for i, p := range processList {
		if i < showNumProcess || showNumProcess == 0 {
			if len(p.Comm) > LIMIT_COMMAND_CHAR {
				p.Comm = p.Comm[:LIMIT_COMMAND_CHAR] + "..."
			}
			newLine := fmt.Sprintf("%5.1f %8s %s", p.CpuPercent, ByteCountIEC(p.Rss*BYTE_IEC), p.Comm)
			if p.TotalP > showCountProcess {
				newLine += fmt.Sprintf(" (%d)", p.TotalP)
			}
			lines = append(lines, newLine)
		}
		totalMemory += p.Rss * BYTE_IEC
		totalProcess += p.TotalP
	}

	linesCombine = strings.Join(lines, "\n")
	return
}

func ByteCountIEC(b int) string {
	const unit = BYTE_IEC
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	// exp++
	return fmt.Sprintf("%.1f%cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

func SortProcess(mapProccess map[string]*Process) []*Process {
	processList := []*Process{}
	for _, p := range mapProccess {
		processList = append(processList, p)
	}

	sort.Slice(processList, func(i, j int) bool {
		return processList[i].Rss > processList[j].Rss
	})
	return processList
}
