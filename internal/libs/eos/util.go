package eos

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var cpuLimitExp = regexp.MustCompile(`subjective cpu limit: (\d+), account cpu limit: (\d+)`)
var cpuLimitErrStr = "account cpu limit"

// 檢查是否為 CPU Limit 錯誤，並判斷是否為 CPU 不足，若CPU還有剩餘，則返回true(可能代表節點忙碌)
func IsCPULimitErrorFalsePositive(err error) (bool, error) {
	if strings.Contains(err.Error(), cpuLimitErrStr) {
		matches := cpuLimitExp.FindStringSubmatch(err.Error())
		if len(matches) > 2 {
			subjectiveCPU := matches[1]
			accountCPULimit := matches[2]

			subjectiveCPUInt, err := strconv.Atoi(subjectiveCPU)
			if err != nil {
				return false, fmt.Errorf("subjectiveCPU strconv.Atoi error: %s", err)
			}

			accountCPULimitInt, err := strconv.Atoi(accountCPULimit)
			if err != nil {
				return false, fmt.Errorf("accountCPULimit strconv.Atoi error: %s", err)
			}

			if subjectiveCPUInt < accountCPULimitInt {
				return true, nil
			}

		} else {
			fmt.Println("No matches found")
		}
	}
	return false, nil
}
